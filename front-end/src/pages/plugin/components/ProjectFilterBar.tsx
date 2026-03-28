import React from 'react';
import { Button, Card, Input, Select, Segmented, Space } from 'antd';
import { AppstoreOutlined, PlusOutlined, SearchOutlined, UnorderedListOutlined } from '@ant-design/icons';

type Option = { label: string; value: string };
type ProjectStatusFilter = 'all' | 'planning' | 'active' | 'offlined';

type ProjectFilterBarProps = {
  keyword: string;
  onKeywordChange: (value: string) => void;
  statusFilter: ProjectStatusFilter;
  onStatusChange: (value: ProjectStatusFilter) => void;
  ownerFilter: string;
  onOwnerChange: (value: string) => void;
  viewMode: 'card' | 'list';
  onViewModeChange: (value: 'card' | 'list') => void;
  statusOptions: Option[];
  ownerOptions: Option[];
  canCreateProject: boolean;
  onCreateProject: () => void;
};

const filterBarStyle: React.CSSProperties = {
  borderRadius: 8,
  border: '1px solid #e5e6eb',
  boxShadow: '0 1px 2px rgba(15, 23, 42, 0.04)',
};

const ProjectFilterBar: React.FC<ProjectFilterBarProps> = ({
  keyword,
  onKeywordChange,
  statusFilter,
  onStatusChange,
  ownerFilter,
  onOwnerChange,
  viewMode,
  onViewModeChange,
  statusOptions,
  ownerOptions,
  canCreateProject,
  onCreateProject,
}) => {
  return (
    <Card bordered={false} style={filterBarStyle}>
      <Space wrap style={{ width: '100%', justifyContent: 'space-between' }}>
        <Space wrap size={12}>
          <Input
            allowClear
            prefix={<SearchOutlined />}
            placeholder="搜索项目名称、编码、负责人或仓库地址"
            style={{ width: 300 }}
            value={keyword}
            onChange={(event) => onKeywordChange(event.target.value)}
          />
          <Select style={{ width: 150 }} value={statusFilter} onChange={onStatusChange} options={statusOptions} />
          <Select style={{ width: 150 }} value={ownerFilter} onChange={onOwnerChange} options={ownerOptions} />
        </Space>
        <Space>
          <Segmented
            value={viewMode}
            onChange={(value) => onViewModeChange(value as 'card' | 'list')}
            options={[
              { label: <AppstoreOutlined />, value: 'card' },
              { label: <UnorderedListOutlined />, value: 'list' },
            ]}
          />
          {canCreateProject ? (
            <Button type="primary" icon={<PlusOutlined />} onClick={onCreateProject}>
              新建项目
            </Button>
          ) : null}
        </Space>
      </Space>
    </Card>
  );
};

export default ProjectFilterBar;
