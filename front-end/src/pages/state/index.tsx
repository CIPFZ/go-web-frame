import React, { useEffect, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import { ProCard } from '@ant-design/pro-components';
import { Alert, Descriptions, Progress, Space, Statistic, Tag } from 'antd';
import { getServerState } from '@/services/api/state';

function formatBytes(bytes: number, decimals = 2): string {
  if (!bytes) return '0 Bytes';
  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(dm))} ${sizes[i]}`;
}

function calculatePercent(used: number, total: number): number {
  if (!total) return 0;
  return parseFloat(((used / total) * 100).toFixed(2));
}

const ServerStatusPage: React.FC = () => {
  const [loading, setLoading] = useState(true);
  const [serverInfo, setServerInfo] = useState<API.ServerInfo | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchServerInfo = async (isFirstLoad = false) => {
      if (isFirstLoad) setLoading(true);
      setError(null);
      try {
        const response = await getServerState();
        if (response.code === 0) {
          setServerInfo(response.data.server);
        } else {
          setError(response.msg || '获取服务器信息失败');
        }
      } catch (e: any) {
        setError(e?.message || '网络请求错误');
      } finally {
        if (isFirstLoad) setLoading(false);
      }
    };

    fetchServerInfo(true);
    const timerId = setInterval(() => fetchServerInfo(false), 10000);
    return () => clearInterval(timerId);
  }, []);

  const renderContent = () => {
    if (error) {
      return <Alert message="加载失败" description={error} type="error" showIcon />;
    }
    if (!serverInfo) return null;

    const { os, cpu, ram, disk, io } = serverInfo;
    const avgCpuUsage = cpu.cpus?.length ? cpu.cpus.reduce((a, b) => a + b, 0) / cpu.cpus.length : 0;
    const ramUsagePercent = calculatePercent(ram.used, ram.total);
    const logicalCores = cpu.cpus?.length || os.numCpu || 0;
    const physicalCores = cpu.cores || 0;

    return (
      <ProCard gutter={[16, 16]} wrap>
        <ProCard colSpan={24} title="系统信息">
          <Descriptions bordered size="small">
            <Descriptions.Item label="操作系统">
              <Tag color="blue">{os.goos}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="Go 版本">{os.goVersion}</Descriptions.Item>
            <Descriptions.Item label="Compiler">{os.compiler}</Descriptions.Item>
            <Descriptions.Item label="逻辑核心">{logicalCores}</Descriptions.Item>
            <Descriptions.Item label="物理核心">{physicalCores}</Descriptions.Item>
            <Descriptions.Item label="Goroutines">
              <Statistic value={os.numGoroutine} suffix="个" valueStyle={{ fontSize: '1rem' }} />
            </Descriptions.Item>
          </Descriptions>
        </ProCard>

        <ProCard colSpan={{ xs: 24, md: 12 }} title="CPU 状态">
          <Progress type="dashboard" percent={parseFloat(avgCpuUsage.toFixed(2))} format={(p) => `${p}%`} />
          <Statistic title="平均使用率" value={avgCpuUsage} precision={2} suffix="%" style={{ marginLeft: 24 }} />
          <Descriptions size="small" column={1} style={{ marginTop: 16 }}>
            <Descriptions.Item label="Load 1">{(cpu.load1 || 0).toFixed(2)}</Descriptions.Item>
            <Descriptions.Item label="Load 5">{(cpu.load5 || 0).toFixed(2)}</Descriptions.Item>
            <Descriptions.Item label="Load 15">{(cpu.load15 || 0).toFixed(2)}</Descriptions.Item>
          </Descriptions>
        </ProCard>

        <ProCard colSpan={{ xs: 24, md: 12 }} title="内存使用率">
          <Statistic title="已使用" value={formatBytes(ram.used)} />
          <Statistic title="总共" value={formatBytes(ram.total)} />
          <Progress percent={ramUsagePercent} />
        </ProCard>

        <ProCard colSpan={24} title="磁盘空间与 IO">
          <Descriptions size="small" column={2} style={{ marginBottom: 12 }}>
            <Descriptions.Item label="总读取">{formatBytes(io?.readBytes || 0)}</Descriptions.Item>
            <Descriptions.Item label="总写入">{formatBytes(io?.writeBytes || 0)}</Descriptions.Item>
          </Descriptions>
          <Space direction="vertical" style={{ width: '100%' }}>
            {disk.map((d, index) => {
              const diskUsagePercent = calculatePercent(d.used, d.total);
              return (
                <div key={index}>
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <strong>{d.mountPoint}</strong>
                    <span>{`${formatBytes(d.used)} / ${formatBytes(d.total)}`}</span>
                  </div>
                  <Progress percent={diskUsagePercent} />
                  <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 12, color: '#666' }}>
                    <span>Read: {formatBytes(d.readBytes || 0)}</span>
                    <span>Write: {formatBytes(d.writeBytes || 0)}</span>
                  </div>
                </div>
              );
            })}
          </Space>
        </ProCard>
      </ProCard>
    );
  };

  return (
    <PageContainer title="服务器状态" loading={loading}>
      {renderContent()}
    </PageContainer>
  );
};

export default ServerStatusPage;
