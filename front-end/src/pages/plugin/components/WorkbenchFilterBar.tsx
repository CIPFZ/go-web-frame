import React from 'react';
import { Card, Input, Select, Space } from 'antd';
import { SearchOutlined } from '@ant-design/icons';

type FilterOption = {
  label: React.ReactNode;
  value: string;
};

type WorkbenchFilterBarProps = {
  keyword: string;
  onKeywordChange: (value: string) => void;
  requestType: string;
  onRequestTypeChange: (value: string) => void;
  status: string;
  onStatusChange: (value: string) => void;
  requestTypeOptions: FilterOption[];
  statusOptions: FilterOption[];
};

const filterBarStyle: React.CSSProperties = {
  borderRadius: 8,
  border: '1px solid #e5e6eb',
  boxShadow: '0 1px 2px rgba(15, 23, 42, 0.04)',
};

const WorkbenchFilterBar: React.FC<WorkbenchFilterBarProps> = ({
  keyword,
  onKeywordChange,
  requestType,
  onRequestTypeChange,
  status,
  onStatusChange,
  requestTypeOptions,
  statusOptions,
}) => {
  return (
    <Card bordered={false} bodyStyle={{ padding: 16 }} style={filterBarStyle}>
      <Space wrap style={{ width: '100%', justifyContent: 'space-between' }}>
        <Space wrap size={12}>
          <Input
            allowClear
            prefix={<SearchOutlined />}
            placeholder="搜索项目、版本号或关键字"
            style={{ width: 280 }}
            value={keyword}
            onChange={(event) => onKeywordChange(event.target.value)}
          />
          <Select
            style={{ width: 180 }}
            value={requestType}
            onChange={onRequestTypeChange}
            options={requestTypeOptions}
          />
          <Select
            style={{ width: 180 }}
            value={status}
            onChange={onStatusChange}
            options={statusOptions}
          />
        </Space>
      </Space>
    </Card>
  );
};

export default WorkbenchFilterBar;
