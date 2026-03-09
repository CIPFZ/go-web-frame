import React, { useRef, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import { ProTable } from '@ant-design/pro-components';
import type { ProColumns, ActionType } from '@ant-design/pro-components';
import { Button, Tag, Space, message, Popconfirm, Modal, Typography } from 'antd';
import { DeleteOutlined, EyeOutlined } from '@ant-design/icons';
import { getOperationLogList, deleteOperationLogByIds } from '@/services/api/operationLog';

const { Text } = Typography;

// 日志项类型定义
type OperationLogItem = {
  ID: number;
  CreatedAt: string;
  ip: string;
  method: string;
  path: string;
  status: number;
  latency: number; // ns
  agent: string;
  body: string;
  resp: string;
  user?: { nickName: string; userName: string };
  traceId?: string;
  error_msg?: string;
};

// 格式化 JSON 辅助函数
const prettyJson = (str: string) => {
  try {
    return JSON.stringify(JSON.parse(str), null, 2);
  } catch (e) {
    return str;
  }
};

const OperationLogTable: React.FC = () => {
  const actionRef = useRef<ActionType>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [currentRow, setCurrentRow] = useState<OperationLogItem>();
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);

  // 批量删除
  const handleBatchDelete = async () => {
    if (!selectedRowKeys.length) return;
    try {
      const res = await deleteOperationLogByIds({ ids: selectedRowKeys as number[] });
      if (res.code === 0) {
        message.success('删除成功');
        setSelectedRowKeys([]);
        actionRef.current?.reload();
      } else {
        message.error(res.msg || '删除失败');
      }
    } catch (error) {
      message.error('请求出错');
    }
  };

  const columns: ProColumns<OperationLogItem>[] = [
    {
      title: 'ID',
      dataIndex: 'ID',
      width: 60,
      search: false,
    },
    {
      title: '操作人',
      dataIndex: 'user_id',
      width: 120, // 固定宽度
      ellipsis: true, // 超出显示省略号
      render: (_, record) => (
        <Tag color="blue" style={{ maxWidth: '100%', overflow: 'hidden', textOverflow: 'ellipsis' }}>
          {record.user?.nickName || record.user?.userName || 'Unknown'}
        </Tag>
      ),
    },
    {
      title: '日期',
      dataIndex: 'created_at',
      valueType: 'dateTimeRange',
      hideInTable: true, // 只在搜索栏显示
      search: {
        transform: (value) => ({ startDate: value[0], endDate: value[1] }),
      },
    },
    {
      title: '操作时间',
      dataIndex: 'CreatedAt',
      valueType: 'dateTime',
      search: false, // 表格中只展示时间点
      width: 160,
    },
    // ✨ 2. 优化状态列 (改为胶囊 Tag)
    {
      title: '状态',
      dataIndex: 'status',
      width: 80,
      align: 'center',
      valueEnum: {
        200: { text: '200', status: 'Success' },
        400: { text: '400', status: 'Warning' },
        401: { text: '401', status: 'Warning' },
        403: { text: '403', status: 'Error' },
        500: { text: '500', status: 'Error' },
      },
      // 自定义渲染为 Tag
      render: (_, record) => {
        let color = 'default';
        if (record.status === 200) color = 'success';
        else if (record.status >= 400 && record.status < 500) color = 'warning';
        else if (record.status >= 500) color = 'error';

        return <Tag color={color}>{record.status}</Tag>;
      }
    },
    {
      title: '耗时',
      dataIndex: 'latency',
      width: 100,
      search: false,
      render: (val) => {
        const ms = Number(val) / 1000000;
        let color = 'green';
        if (ms > 500) color = 'orange';
        if (ms > 1000) color = 'red';
        return <Tag color={color}>{ms.toFixed(2)} ms</Tag>;
      },
    },
    {
      title: '方法',
      dataIndex: 'method',
      width: 80,
      valueEnum: {
        GET: { text: 'GET', status: 'Default' }, // 理论上我们过滤了 GET
        POST: { text: 'POST', status: 'Processing' },
        PUT: { text: 'PUT', status: 'Warning' },
        DELETE: { text: 'DELETE', status: 'Error' },
      },
    },
    {
      title: '请求路径',
      dataIndex: 'path',
      copyable: true,
      ellipsis: true,
    },
    {
      title: '请求IP',
      dataIndex: 'ip',
      width: 120,
      copyable: true,
    },
    {
      title: 'TraceID', // OTel 链路追踪
      dataIndex: 'traceId',
      copyable: true,
      ellipsis: true,
      search: false, // 可以开启搜索，方便排错
    },
    {
      title: '操作',
      valueType: 'option',
      fixed: 'right',
      width: 80,
      render: (_, record) => (
        <a onClick={() => { setCurrentRow(record); setIsModalOpen(true); }}>
          <EyeOutlined /> 详情
        </a>
      ),
    },
  ];

  return (
    <PageContainer title={false}>
      <ProTable<OperationLogItem>
        headerTitle={false}
        actionRef={actionRef}
        rowKey="ID"
        rowSelection={{
          selectedRowKeys,
          onChange: (keys) => setSelectedRowKeys(keys),
        }}
        request={async (params) => {
          const res = await getOperationLogList({
            page: params.current,
            pageSize: params.pageSize,
            ...params,
            // 处理时间范围搜索参数映射
            startDate: params.created_at?.[0],
            endDate: params.created_at?.[1],
          });
          return {
            data: res.data?.list || [],
            success: res.code === 0,
            total: res.data?.total || 0,
          };
        }}
        columns={columns}
        scroll={{ x: 1300 }}
        toolBarRender={() => [
          selectedRowKeys.length > 0 && (
            <Popconfirm
              key="batchDelete"
              title={`确定删除选中的 ${selectedRowKeys.length} 条日志吗？`}
              onConfirm={handleBatchDelete}
            >
              <Button danger type="primary">
                <DeleteOutlined /> 批量删除
              </Button>
            </Popconfirm>
          ),
        ]}
      />

      {/* 详情模态框 */}
      <Modal
        title="请求详情"
        width={800}
        open={isModalOpen}
        onCancel={() => setIsModalOpen(false)}
        footer={null}
      >
        {currentRow && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            <div>
              <Text strong>请求 Body:</Text>
              <div style={{
                marginTop: 8,
                padding: 12,
                background: '#f5f5f5',
                borderRadius: 4,
                maxHeight: 300,
                overflow: 'auto',
                whiteSpace: 'pre-wrap',
                fontFamily: 'monospace'
              }}>
                {prettyJson(currentRow.body)}
              </div>
            </div>

            <div>
              <Text strong>响应 Body:</Text>
              <div style={{
                marginTop: 8,
                padding: 12,
                background: '#f5f5f5',
                borderRadius: 4,
                maxHeight: 300,
                overflow: 'auto',
                whiteSpace: 'pre-wrap',
                fontFamily: 'monospace'
              }}>
                {prettyJson(currentRow.resp)}
              </div>
            </div>

            {currentRow.error_msg && (
              <div>
                <Text strong type="danger">错误信息:</Text>
                <div style={{ marginTop: 8, padding: 8, background: '#fff1f0', border: '1px solid #ffccc7' }}>
                  {currentRow.error_msg}
                </div>
              </div>
            )}
          </div>
        )}
      </Modal>
    </PageContainer>
  );
};

export default OperationLogTable;
