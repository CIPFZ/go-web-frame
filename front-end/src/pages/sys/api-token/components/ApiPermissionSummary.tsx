import React from 'react';
import { Button, Popover, Space, Tag, Typography } from 'antd';

import type { ApiTokenItem } from '@/services/api/apiToken';
import { buildPermissionSummary } from '../helpers';

type ApiPermissionSummaryProps = {
  apis?: ApiTokenItem['apis'];
  maxVisible?: number;
};

export const ApiPermissionSummary: React.FC<ApiPermissionSummaryProps> = ({
  apis,
  maxVisible = 2,
}) => {
  const summary = buildPermissionSummary(apis, maxVisible);

  if (!summary.total) {
    return <Typography.Text type="secondary">未授权</Typography.Text>;
  }

  return (
    <Space direction="vertical" size={6} style={{ width: '100%' }}>
      <div data-testid="api-permission-count-row">
        <Tag color="geekblue">{`${summary.total} 个 API`}</Tag>
      </div>

      <Space
        direction="vertical"
        size={6}
        style={{ width: '100%' }}
        data-testid="api-permission-visible-list"
      >
        {summary.visible.map((item) => (
          <div key={item.key} data-testid="api-permission-visible-api-row">
            <Tag color="blue" style={{ marginInlineEnd: 0, maxWidth: '100%' }}>
              {item.label}
            </Tag>
          </div>
        ))}

        {summary.hidden.length > 0 ? (
          <div data-testid="api-permission-overflow-row">
            <Space size={8} wrap>
              <Tag style={{ marginInlineEnd: 0 }}>{`+${summary.hidden.length}`}</Tag>
              <Popover
                trigger="click"
                placement="bottomLeft"
                content={
                  <div
                    data-testid="api-permission-popover-list"
                    style={{
                      maxHeight: 240,
                      overflowY: 'auto',
                      paddingRight: 4,
                    }}
                  >
                    <Space direction="vertical" size={8}>
                      {summary.hidden.map((item) => (
                        <div key={item.key}>
                          <Tag color="processing" style={{ marginInlineEnd: 0, maxWidth: '100%' }}>
                            {item.label}
                          </Tag>
                        </div>
                      ))}
                    </Space>
                  </div>
                }
              >
                <Button type="link" size="small">
                  查看全部
                </Button>
              </Popover>
            </Space>
          </div>
        ) : null}
      </Space>
    </Space>
  );
};
