package extlogin

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// TicketTTL 是回调发出的一次性票据的有效期：前端在这段时间内拿它换登录凭证。
const TicketTTL = time.Minute

// ErrTicketInvalid 表示票据不存在、已用过或已过期。
var ErrTicketInvalid = errors.New("extlogin: ticket is invalid")

// Tickets 保存回调发出的一次性票据（换到登录凭证即作废）。登录凭证不出现在回调地址里，只出现在前端换取的响应中。
type Tickets interface {
	Issue(ctx context.Context, userID uint) (string, error)
	Redeem(ctx context.Context, ticket string) (uint, error)
}

func newTicket() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}

// NewTickets 有 Redis 时用 Redis（多实例共享），否则用进程内存（只适合单实例）。
func NewTickets(client redis.UniversalClient) Tickets {
	if client != nil {
		return &redisTickets{client: client}
	}
	return &memoryTickets{tickets: map[string]memoryTicket{}, now: time.Now}
}

type redisTickets struct{ client redis.UniversalClient }

const ticketPrefix = "extlogin:ticket:"

func (t *redisTickets) Issue(ctx context.Context, userID uint) (string, error) {
	ticket, err := newTicket()
	if err != nil {
		return "", err
	}
	return ticket, t.client.Set(ctx, ticketPrefix+ticket, userID, TicketTTL).Err()
}

func (t *redisTickets) Redeem(ctx context.Context, ticket string) (uint, error) {
	v, err := t.client.GetDel(ctx, ticketPrefix+ticket).Result()
	if errors.Is(err, redis.Nil) {
		return 0, ErrTicketInvalid
	}
	if err != nil {
		return 0, err
	}
	id, err := strconv.ParseUint(v, 10, 64)
	if err != nil || id == 0 {
		return 0, ErrTicketInvalid
	}
	return uint(id), nil
}

type memoryTicket struct {
	userID  uint
	expires time.Time
}

type memoryTickets struct {
	mu      sync.Mutex
	tickets map[string]memoryTicket
	now     func() time.Time
}

func (t *memoryTickets) Issue(_ context.Context, userID uint) (string, error) {
	ticket, err := newTicket()
	if err != nil {
		return "", err
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	now := t.now()
	for k, v := range t.tickets {
		if now.After(v.expires) {
			delete(t.tickets, k)
		}
	}
	t.tickets[ticket] = memoryTicket{userID: userID, expires: now.Add(TicketTTL)}
	return ticket, nil
}

func (t *memoryTickets) Redeem(_ context.Context, ticket string) (uint, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	v, ok := t.tickets[ticket]
	delete(t.tickets, ticket)
	if !ok || t.now().After(v.expires) {
		return 0, ErrTicketInvalid
	}
	return v.userID, nil
}
