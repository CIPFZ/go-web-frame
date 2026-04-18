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
    <Space size={[6, 6]} wrap>
      <Tag color="geekblue">{`${summary.total} 个 API`}</Tag>
      {summary.visible.map((item) => (
        <Tag key={item.key} color="blue">
          {item.label}
        </Tag>
      ))}
      {summary.hidden.length > 0 ? (
        <Popover
          trigger="click"
          placement="bottomLeft"
          content={
            <Space direction="vertical" size={8}>
              {summary.hidden.map((item) => (
                <Tag key={item.key} color="processing">
                  {item.label}
                </Tag>
              ))}
            </Space>
          }
        >
          <Button type="link" size="small">
            查看全部
          </Button>
        </Popover>
      ) : null}
      {summary.hidden.length > 0 ? <Tag>{`+${summary.hidden.length}`}</Tag> : null}
    </Space>
  );
};
