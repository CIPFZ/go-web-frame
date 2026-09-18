import React, { useEffect, useState } from 'react';
import { useIntl } from '@umijs/max';
import { PageContainer } from '@ant-design/pro-components';
import { Alert, Button, Card, Col, Descriptions, Form, Image, Input, Row, Space, Table, Tag, Typography } from 'antd';
import { FileOutlined, FileImageOutlined, FileTextOutlined, FileZipOutlined, HddOutlined, PlayCircleOutlined, SoundOutlined, SearchOutlined } from '@ant-design/icons';
import { loadCover, previewMagnet, type MagnetPreview, type ResourceType } from '@/services/magnet';

const icons: Record<ResourceType, React.ReactNode> = {
  video: <PlayCircleOutlined />, audio: <SoundOutlined />, image: <FileImageOutlined />,
  document: <FileTextOutlined />, archive: <FileZipOutlined />, disk_image: <HddOutlined />, other: <FileOutlined />,
};


export default function MagnetPreviewPage() {
  const intl = useIntl();
  const t = {
    title: intl.formatMessage({id: 'magnet.title'}),
    intro: intl.formatMessage({id: 'magnet.intro'}),
    input: intl.formatMessage({id: 'magnet.input'}),
    required: intl.formatMessage({id: 'magnet.required'}),
    submit: intl.formatMessage({id: 'magnet.submit'}),
    pending: intl.formatMessage({id: 'magnet.pending'}),
    error: intl.formatMessage({id: 'magnet.error'}),
    empty: intl.formatMessage({id: 'magnet.empty'}),
    noCover: intl.formatMessage({id: 'magnet.noCover'}),
    size: intl.formatMessage({id: 'magnet.size'}),
    count: intl.formatMessage({id: 'magnet.count'}),
    type: intl.formatMessage({id: 'magnet.type'}),
    files: intl.formatMessage({id: 'magnet.files'}),
    name: intl.formatMessage({id: 'magnet.name'}),
    fileSize: intl.formatMessage({id: 'magnet.fileSize'}),
    search: intl.formatMessage({id: 'magnet.search'}),
    truncated: intl.formatMessage({id: 'magnet.truncated'}),
    cached: intl.formatMessage({id: 'magnet.cached'}),
    hash: intl.formatMessage({id: 'magnet.hash'}),
    retrieved: intl.formatMessage({id: 'magnet.retrieved'}),
    types: {
      video: intl.formatMessage({id: 'magnet.types.video'}),
      audio: intl.formatMessage({id: 'magnet.types.audio'}),
      image: intl.formatMessage({id: 'magnet.types.image'}),
      document: intl.formatMessage({id: 'magnet.types.document'}),
      archive: intl.formatMessage({id: 'magnet.types.archive'}),
      disk_image: intl.formatMessage({id: 'magnet.types.disk_image'}),
      other: intl.formatMessage({id: 'magnet.types.other'})
    },
  };
  const [result, setResult] = useState<MagnetPreview>();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [filter, setFilter] = useState('');
  const [coverURL, setCoverURL] = useState('');
  const [coverFailed, setCoverFailed] = useState(false);

  useEffect(() => {
    let active = true;
    let objectURL = '';
    setCoverURL('');
    setCoverFailed(false);
    if (result?.cover?.url) {
      loadCover(result.cover.url).then(blob => {
        if (active && blob.type.startsWith('image/')) {
          objectURL = URL.createObjectURL(blob);
          setCoverURL(objectURL);
        }
      }).catch(() => { if (active) setCoverFailed(true); });
    }
    return () => { active = false; if (objectURL) URL.revokeObjectURL(objectURL); };
  }, [result]);

  async function submit(values: { magnet: string }) {
    setLoading(true); setError(''); setResult(undefined); setFilter('');
    try {
      const response = await previewMagnet(values.magnet.trim());
      if (response.code !== 0) { setError(response.msg || t.error); return; }
      setResult(response.data);
    } catch { setError(t.error); }
    finally { setLoading(false); }
  }

  const kind = result?.content_type || 'other';
  return (
    <PageContainer title={t.title}>
      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        <Card>
          <Typography.Paragraph type="secondary">{t.intro}</Typography.Paragraph>
          <Form layout="vertical" onFinish={submit}>
            <Form.Item name="magnet" label={t.input} rules={[{ required: true, whitespace: true, message: t.required },
              { pattern: /^\s*magnet:\?/i, message: t.required }]}>
              <Input.TextArea rows={3} maxLength={16384} placeholder="magnet:?xt=urn:btih:…" disabled={loading} />
            </Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} icon={<SearchOutlined />}>{t.submit}</Button>
          </Form>
        </Card>
        {loading && <Alert type="info" showIcon message={t.pending} />}
        {error && <Alert type="error" showIcon message={error} />}
        {result ? <>
          <Card>
            <Row gutter={[24, 24]} align="middle">
              <Col xs={24} sm={6}>
                {coverURL && !coverFailed ? <Image src={coverURL} alt={result.name} style={{ maxHeight: 240, objectFit: 'contain' }} onError={() => setCoverFailed(true)} /> :
                  <div style={{ textAlign: 'center', padding: 24, background: 'rgba(127,127,127,0.06)', borderRadius: 12 }}>
                    <div style={{ fontSize: 64, color: '#1677ff' }}>{icons[kind] || icons.other}</div>
                    <Typography.Text type="secondary">{t.noCover}</Typography.Text>
                  </div>}
              </Col>
              <Col xs={24} sm={18}>
                <Typography.Title level={3} style={{ overflowWrap: 'anywhere' }}>{result.name}</Typography.Title>
                <Space wrap style={{ marginBottom: 16 }}>
                  <Tag color="blue">{t.types[kind] || t.types.other}</Tag>
                  {result.cached && <Tag>{t.cached}</Tag>}
                </Space>
                <Descriptions column={{ xs: 1, sm: 2 }}>
                  <Descriptions.Item label={t.size}>{result.total_size_human}</Descriptions.Item>
                  <Descriptions.Item label={t.count}>{result.file_count}</Descriptions.Item>
                  <Descriptions.Item label={t.hash} span={2}><Typography.Text copyable style={{ overflowWrap: 'anywhere' }}>{result.info_hash}</Typography.Text></Descriptions.Item>
                  <Descriptions.Item label={t.retrieved} span={2}>{new Date(result.retrieved_at).toLocaleString(intl.locale)}</Descriptions.Item>
                </Descriptions>
              </Col>
            </Row>
          </Card>
          <Card title={t.files}>
            <Space direction="vertical" size="middle" style={{ width: '100%' }}>
              {result.files_truncated && <Alert type="warning" showIcon message={t.truncated} />}
              <Input allowClear prefix={<SearchOutlined />} placeholder={t.search} value={filter} onChange={e => setFilter(e.target.value)} />
              <Table rowKey="index" size="small" scroll={{ x: 560 }} pagination={{ pageSize: 20, showSizeChanger: false }}
                dataSource={result.files.filter(file => file.path.toLowerCase().includes(filter.toLowerCase()))}
                columns={[
                  { title: t.name, dataIndex: 'path', render: (path: string) => <span style={{ overflowWrap: 'anywhere' }}>{path}</span> },
                  { title: t.type, dataIndex: 'type', width: 160, filters: Object.entries(t.types).map(([value, text]) => ({ text, value })),
                    onFilter: (value, file) => file.type === value,
                    render: (type: ResourceType) => <Space>{icons[type] || icons.other}{t.types[type] || t.types.other}</Space> },
                  { title: t.fileSize, dataIndex: 'size_human', width: 110, sorter: (a, b) => a.size - b.size },
                ]} />
            </Space>
          </Card>
        </> : !loading && !error && <Card><Typography.Text type="secondary">{t.empty}</Typography.Text></Card>}
      </Space>
    </PageContainer>
  );
}
