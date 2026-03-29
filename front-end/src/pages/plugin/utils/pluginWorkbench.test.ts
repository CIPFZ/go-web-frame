import {
  buildReleaseAssignmentPayload,
  buildReleaseTransitionPayload,
  buildReviewWorkbenchBaseQuery,
  buildPublishWorkbenchQuery,
  resolveDetailEntryMeta,
  resolveInitialTab,
} from './pluginWorkbench';

describe('plugin workbench helpers', () => {
  it('defaults the publish workbench query to assigned items', () => {
    expect(
      buildPublishWorkbenchQuery({
        page: 2,
        pageSize: 20,
        keyword: '  content  ',
        requestType: 'initial',
        status: 'approved',
        scope: 'assigned',
        currentUserId: 15,
      }),
    ).toEqual({
      page: 2,
      pageSize: 20,
      keyword: 'content',
      requestType: 'initial',
      status: 'approved',
      publisherId: 15,
    });
  });

  it('allows the publish workbench to switch to all pending items', () => {
    expect(
      buildPublishWorkbenchQuery({
        page: 1,
        pageSize: 10,
        keyword: '   ',
        requestType: 'all',
        status: 'approved',
        scope: 'all',
        currentUserId: 15,
      }),
    ).toEqual({
      page: 1,
      pageSize: 10,
      keyword: undefined,
      requestType: undefined,
      status: 'approved',
      publisherId: undefined,
    });
  });

  it('scopes the review workbench query only when the assigned view is selected', () => {
    expect(
      buildReviewWorkbenchBaseQuery({
        keyword: '  plugin  ',
        requestType: 'maintenance',
        scope: 'assigned',
        currentUserId: 7,
      }),
    ).toEqual({
      keyword: 'plugin',
      requestType: 'maintenance',
      reviewerId: 7,
    });

    expect(
      buildReviewWorkbenchBaseQuery({
        keyword: '',
        requestType: 'all',
        scope: 'all',
        currentUserId: 7,
      }),
    ).toEqual({
      keyword: undefined,
      requestType: undefined,
      reviewerId: undefined,
    });
  });

  it('only sends publisher assignment when approving a release', () => {
    expect(
      buildReleaseTransitionPayload({
        id: 8,
        action: 'submit_review',
        reviewComment: '  ready  ',
        reviewerId: 21,
      }),
    ).toEqual({
      id: 8,
      action: 'submit_review',
      reviewComment: 'ready',
      reviewerId: 21,
      publisherId: undefined,
    });

    expect(
      buildReleaseTransitionPayload({
        id: 8,
        action: 'approve',
        reviewComment: '  looks good  ',
        publisherId: 12,
      }),
    ).toEqual({
      id: 8,
      action: 'approve',
      reviewComment: 'looks good',
      reviewerId: undefined,
      publisherId: 12,
    });

    expect(
      buildReleaseTransitionPayload({
        id: 8,
        action: 'reject',
        reviewComment: '  fix package  ',
        publisherId: 12,
      }),
    ).toEqual({
      id: 8,
      action: 'reject',
      reviewComment: 'fix package',
      reviewerId: undefined,
      publisherId: undefined,
    });
  });

  it('trims assignment payloads and omits empty assignees', () => {
    expect(
      buildReleaseAssignmentPayload({
        id: 9,
        reviewerId: 18,
        comment: '  switch reviewer  ',
      }),
    ).toEqual({
      id: 9,
      reviewerId: 18,
      publisherId: undefined,
      comment: 'switch reviewer',
    });

    expect(
      buildReleaseAssignmentPayload({
        id: 9,
        reviewerId: 0,
        publisherId: 12,
        comment: '   ',
      }),
    ).toEqual({
      id: 9,
      reviewerId: undefined,
      publisherId: 12,
      comment: undefined,
    });
  });

  it('defaults the detail tab based on the entry source', () => {
    expect(resolveInitialTab(undefined, undefined)).toBe('overview');
    expect(resolveInitialTab('review', undefined)).toBe('review');
    expect(resolveInitialTab('publish', undefined)).toBe('review');
    expect(resolveInitialTab('publish', 'timeline')).toBe('timeline');
  });

  it('returns the correct back navigation metadata for the publish workbench', () => {
    expect(resolveDetailEntryMeta('publish')).toMatchObject({
      source: 'publish',
      backTarget: '/plugin/publish-workbench',
      backLabel: '返回发布工作台',
      sectionLabel: '发布工作台',
    });
  });
});
