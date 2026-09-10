import { t, useI18n } from '@/i18n';
import React from 'react';
import { Button, Popover, Space, Tag, Typography } from 'antd';
import type { ApiTokenItem } from '@/services/system/apiToken';
import { buildPermissionSummary } from '../helpers';
type ApiPermissionSummaryProps = {
  apis?: ApiTokenItem['apis'];
  maxVisible?: number;
};
export const ApiPermissionSummary: React.FC<ApiPermissionSummaryProps> = ({
  apis,
  maxVisible = 2,
}) => {
  useI18n();
  const summary = buildPermissionSummary(apis, maxVisible);
  if (!summary.total) {
    return (
      <Typography.Text type="secondary">
        {t('cms.notAuthorized')}
      </Typography.Text>
    );
  }
  return (
    <Space
      direction="vertical"
      size={6}
      style={{
        width: '100%',
      }}
    >
      <div data-testid="api-permission-count-row">
        <Tag color="geekblue">
          {t('cms.apiCount', {
            value0: summary.total,
          })}
        </Tag>
      </div>

      <Space
        direction="vertical"
        size={6}
        style={{
          width: '100%',
        }}
        data-testid="api-permission-visible-list"
      >
        {summary.visible.map((item) => (
          <div key={item.key} data-testid="api-permission-visible-api-row">
            <Tag
              color="blue"
              style={{
                marginInlineEnd: 0,
                maxWidth: '100%',
              }}
            >
              {item.label}
            </Tag>
          </div>
        ))}

        {summary.hidden.length > 0 ? (
          <div data-testid="api-permission-overflow-row">
            <Space size={8} wrap>
              <Tag
                style={{
                  marginInlineEnd: 0,
                }}
              >{`+${summary.hidden.length}`}</Tag>
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
                          <Tag
                            color="processing"
                            style={{
                              marginInlineEnd: 0,
                              maxWidth: '100%',
                            }}
                          >
                            {item.label}
                          </Tag>
                        </div>
                      ))}
                    </Space>
                  </div>
                }
              >
                <Button type="link" size="small">
                  {t('cms.viewAll')}
                </Button>
              </Popover>
            </Space>
          </div>
        ) : null}
      </Space>
    </Space>
  );
};
