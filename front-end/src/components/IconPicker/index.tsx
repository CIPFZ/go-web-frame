import React, { useState, useMemo } from 'react';
import { Input, Popover, Empty, Pagination } from 'antd';
import * as AntdIcons from '@ant-design/icons';
import { SearchOutlined } from '@ant-design/icons';

// 排除非图标组件的属性
const allIcons: { [key: string]: any } = AntdIcons;
const iconKeys = Object.keys(allIcons).filter(
  (key) => typeof allIcons[key] === 'object' && key.endsWith('Outlined'),
);

interface IconPickerProps {
  value?: string;
  onChange?: (value: string) => void;
}

const IconPicker: React.FC<IconPickerProps> = ({ value, onChange }) => {
  const [visible, setVisible] = useState(false);
  const [searchValue, setSearchValue] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const pageSize = 36;

  // 过滤图标
  const filteredIcons = useMemo(() => {
    return iconKeys.filter((key) =>
      key.toLowerCase().includes(searchValue.toLowerCase()),
    );
  }, [searchValue]);

  // 分页逻辑
  const currentIcons = useMemo(() => {
    const start = (currentPage - 1) * pageSize;
    return filteredIcons.slice(start, start + pageSize);
  }, [filteredIcons, currentPage]);

  const handleSelect = (iconKey: string) => {
    if (onChange) {
      onChange(iconKey);
    }
    setVisible(false);
  };

  const content = (
    <div style={{ width: 300 }}> {/* 1. 稍微收窄宽度 */}
      <Input
        placeholder="搜索图标..."
        prefix={<SearchOutlined />}
        value={searchValue}
        onChange={(e) => {
          setSearchValue(e.target.value);
          setCurrentPage(1);
        }}
        allowClear
      />

      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(6, 1fr)',
        gap: 8,
        height: 240,        // 2. 设定固定高度
        overflowY: 'auto',  // 3. 内容过多时出现内部滚动条，而不是撑大弹窗
        alignContent: 'start',
        marginTop: 12,
        marginBottom: 12,
        paddingRight: 4 // 预留一点滚动条空间
      }}>
        {currentIcons.map((key) => {
          const Icon = allIcons[key];
          return (
            <div
              key={key}
              onClick={() => handleSelect(key)}
              title={key}
              style={{
                cursor: 'pointer',
                padding: 6,
                display: 'flex', // 居中图标
                justifyContent: 'center',
                alignItems: 'center',
                border: value === key ? '1px solid #1890ff' : '1px solid #f0f0f0',
                borderRadius: 4,
                fontSize: 22, // 图标稍大一点
                color: value === key ? '#1890ff' : '#595959',
                transition: 'all 0.3s'
              }}
              onMouseEnter={(e) => {e.currentTarget.style.borderColor = '#1890ff'}}
              onMouseLeave={(e) => {e.currentTarget.style.borderColor = value === key ? '#1890ff' : '#f0f0f0'}}
            >
              <Icon />
            </div>
          );
        })}
        {currentIcons.length === 0 && (
          <div style={{ gridColumn: '1 / span 6', textAlign: 'center', paddingTop: 80 }}>
            <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="无搜索结果" />
          </div>
        )}
      </div>

      <div style={{
        textAlign: 'center', // 4. 分页器居中
        borderTop: '1px solid #f0f0f0',
        paddingTop: 12
      }}>
        <Pagination
          simple
          current={currentPage}
          pageSize={pageSize}
          total={filteredIcons.length}
          onChange={setCurrentPage}
          size="small"
          showSizeChanger={false} // 5. 关键修复：禁用每页条数选择器，防止撑爆宽度
        />
      </div>
    </div>
  );

  return (
    <Popover
      content={content}
      trigger="click"
      open={visible}
      onOpenChange={setVisible}
      placement="bottomLeft"
      destroyTooltipOnHide // 关闭时销毁，节省性能
      zIndex={1050} // 确保在 Modal 之上 (Modal 默认 1000)
    >
      <Input
        value={value}
        readOnly
        placeholder="点击选择图标"
        prefix={value && allIcons[value] ? React.createElement(allIcons[value]) : <SearchOutlined style={{color:'#bfbfbf'}} />}
        onClick={() => setVisible(true)}
        style={{ cursor: 'pointer' }}
      />
    </Popover>
  );
};

export default IconPicker;
