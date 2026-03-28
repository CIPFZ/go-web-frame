import React from 'react';
import { Avatar, Card, Space, Tag, Typography } from 'antd';

type ProjectSummaryCardProps = {
  layout?: 'card' | 'list';
  code: string;
  nameZh: string;
  nameEn?: string;
  latestVersion: string;
  workflowSummary: string;
  statusLabel: string;
  statusColor: string;
  onClick?: () => void;
};

const baseStyle: React.CSSProperties = {
  borderRadius: 8,
  border: '1px solid #e5e6eb',
  boxShadow: '0 1px 2px rgba(15, 23, 42, 0.04)',
};

const ProjectSummaryCard: React.FC<ProjectSummaryCardProps> = ({
  layout = 'card',
  code,
  nameZh,
  nameEn,
  latestVersion,
  workflowSummary,
  statusLabel,
  statusColor,
  onClick,
}) => {
  const isList = layout === 'list';

  return (
    <Card
      bordered={false}
      hoverable={!!onClick}
      style={{
        ...baseStyle,
        height: isList ? 'auto' : '100%',
        cursor: onClick ? 'pointer' : 'default',
      }}
      bodyStyle={{
        padding: isList ? 16 : 14,
        display: 'flex',
        flexDirection: 'column',
        height: '100%',
      }}
      onClick={onClick}
    >
      <Space align="start" size={12} style={{ width: '100%' }}>
        <Avatar
          shape="square"
          size={isList ? 44 : 48}
          style={{
            background: '#e8f3ff',
            color: '#1677ff',
            borderRadius: 12,
            fontSize: 18,
            fontWeight: 700,
            flex: 'none',
          }}
        >
          {(code || 'P').slice(0, 1).toUpperCase()}
        </Avatar>

        <div style={{ flex: 1, minWidth: 0 }}>
          <Space align="start" style={{ justifyContent: 'space-between', width: '100%' }}>
            <div style={{ minWidth: 0 }}>
              <Typography.Title level={5} style={{ margin: 0, fontSize: 16, lineHeight: 1.4 }} ellipsis>
                {nameZh || '-'}
              </Typography.Title>
              <Typography.Paragraph type="secondary" style={{ marginBottom: 0, fontSize: 12 }} ellipsis>
                {nameEn || '-'}
              </Typography.Paragraph>
              <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                {code || '-'}
              </Typography.Text>
            </div>
            <Tag color={statusColor} style={{ margin: 0, flex: 'none' }}>
              {statusLabel}
            </Tag>
          </Space>

          <Space direction="vertical" size={4} style={{ width: '100%', marginTop: 12 }}>
            <div>
              <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                最新版本
              </Typography.Text>
              <div style={{ fontWeight: 600 }}>{latestVersion || '-'}</div>
            </div>
            <div>
              <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                流程摘要
              </Typography.Text>
              <div style={{ fontWeight: 600 }}>{workflowSummary || '-'}</div>
            </div>
          </Space>

        </div>
      </Space>
    </Card>
  );
};

export default ProjectSummaryCard;
