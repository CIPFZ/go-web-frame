package extlogin

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

type fakeProvider struct{ key string }

func (p fakeProvider) Key() string  { return p.key }
func (p fakeProvider) Name() string { return "Fake " + p.key }
func (p fakeProvider) LoginURL(context.Context, string) (string, error) {
	return "https://sso.example.com/login", nil
}
func (p fakeProvider) Exchange(context.Context, url.Values) (*Identity, error) { return nil, nil }

func TestRegisterChecksKeysAndOptions(t *testing.T) {
	t.Cleanup(reset)
	Register(fakeProvider{"corp-sso"}, Options{})
	Register(fakeProvider{"b"}, Options{AutoCreate: true, RoleID: 888})
	require.Equal(t, []Info{{Key: "b", Name: "Fake b"}, {Key: "corp-sso", Name: "Fake corp-sso"}}, List())
	_, opt, ok := Get("b")
	require.True(t, ok)
	require.EqualValues(t, 888, opt.RoleID)
	require.Panics(t, func() { Register(fakeProvider{"corp-sso"}, Options{}) }, "duplicate")
	require.Panics(t, func() { Register(fakeProvider{"Bad Key"}, Options{}) }, "key format")
	require.Panics(t, func() { Register(fakeProvider{"x"}, Options{AutoCreate: true}) }, "auto-create without a role")
	_, _, ok = Get("missing")
	require.False(t, ok)
}

func ticketSuite(t *testing.T, tickets Tickets) {
	ctx := context.Background()
	ticket, err := tickets.Issue(ctx, 42)
	require.NoError(t, err)
	require.Len(t, ticket, 64)
	id, err := tickets.Redeem(ctx, ticket)
	require.NoError(t, err)
	require.EqualValues(t, 42, id)
	_, err = tickets.Redeem(ctx, ticket)
	require.ErrorIs(t, err, ErrTicketInvalid, "a ticket works once")
	_, err = tickets.Redeem(ctx, "unknown")
	require.ErrorIs(t, err, ErrTicketInvalid)
}

func TestMemoryTicketsAreSingleUseAndExpire(t *testing.T) {
	tickets := NewTickets(nil).(*memoryTickets)
	ticketSuite(t, tickets)
	now := time.Now()
	tickets.now = func() time.Time { return now }
	ticket, err := tickets.Issue(context.Background(), 7)
	require.NoError(t, err)
	tickets.now = func() time.Time { return now.Add(TicketTTL + time.Second) }
	_, err = tickets.Redeem(context.Background(), ticket)
	require.ErrorIs(t, err, ErrTicketInvalid)
}

func TestRedisTicketsAreSingleUseAndExpire(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	ticketSuite(t, NewTickets(client))
	ticket, err := NewTickets(client).Issue(context.Background(), 7)
	require.NoError(t, err)
	server.FastForward(TicketTTL + time.Second)
	_, err = NewTickets(client).Redeem(context.Background(), ticket)
	require.ErrorIs(t, err, ErrTicketInvalid)
}
