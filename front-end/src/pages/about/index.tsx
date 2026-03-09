// src/page/about/index.tsx
import React from 'react';
import { PageContainer, ProCard } from '@ant-design/pro-components';
import {Descriptions, Tag, List, Typography, Space} from 'antd';
import {
  HeartOutlined,
  GithubOutlined,
  BookOutlined,
  ReadOutlined,
  CodeOutlined,
} from '@ant-design/icons';

// 假设的项目信息，你可以替换为你自己的
const projectInfo = {
  name: 'My CMS Admin (Antd Pro Edition)',
  version: '1.0.0',
  description: '一个基于 Go 和 React 的现代化内容管理系统。',
  documentation: '#', // 替换为你的文档链接
  github: '#', // 替换为你的仓库链接
};

// 前端技术栈
const frontendTech = [
  {
    name: 'React',
    desc: '核心 UI 库',
    icon: <ReadOutlined style={{ color: '#61DAFB' }} />,
  },
  {
    name: 'Ant Design Pro',
    desc: '企业级前端 UI 框架',
    icon: <HeartOutlined style={{ color: '#1890ff' }} />,
  },
  {
    name: 'UmiJS',
    desc: '应用路由、构建和开发工具',
    icon: <BookOutlined style={{ color: '#1890ff' }} />,
  },
  {
    name: 'TypeScript',
    desc: '强类型 JavaScript 超集',
    icon: <CodeOutlined style={{ color: '#3178C6' }} />,
  },
];

// 后端技术栈
const backendTech = [
  {
    name: 'Go (Golang)',
    desc: '核心 API 语言',
    icon: <CodeOutlined style={{ color: '#00ADD8' }} />,
  },
  {
    name: 'Gin',
    desc: '高性能 HTTP Web 框架',
    icon: <CodeOutlined style={{ color: '#00ADD8' }} />,
  },
  {
    name: 'GORM',
    desc: 'Go 语言的 ORM 库',
    icon: <CodeOutlined style={{ color: '#00ADD8' }} />,
  },
  {
    name: 'MySQL/PostgreSQL',
    desc: '关系型数据库',
    icon: <CodeOutlined style={{ color: '#00ADD8' }} />,
  },
];

const AboutPage: React.FC = () => {
  return (
    <PageContainer title="关于我们">
      <ProCard
        style={{ marginBottom: 16 }}
        title={
          <Space>
            <GithubOutlined />
            <Typography.Title level={4} style={{ margin: 0 }}>
              {projectInfo.name}
            </Typography.Title>
          </Space>
        }
      >
        <Typography.Paragraph>{projectInfo.description}</Typography.Paragraph>
        <Descriptions bordered size="small">
          <Descriptions.Item label="当前版本">
            <Tag color="blue">{projectInfo.version}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="文档地址">
            <a href={projectInfo.documentation} target="_blank" rel="noopener noreferrer">
              查看文档
            </a>
          </Descriptions.Item>
          <Descriptions.Item label="Github">
            <a href={projectInfo.github} target="_blank" rel="noopener noreferrer">
              项目仓库
            </a>
          </Descriptions.Item>
        </Descriptions>
      </ProCard>

      <ProCard.Group title="技术栈" direction="column">
        <ProCard title="前端技术" colSpan={{ xs: 24, md: 12 }}>
          <List
            itemLayout="horizontal"
            dataSource={frontendTech}
            renderItem={(item) => (
              <List.Item>
                <List.Item.Meta
                  avatar={item.icon}
                  title={item.name}
                  description={item.desc}
                />
              </List.Item>
            )}
          />
        </ProCard>

        <ProCard title="后端技术" colSpan={{ xs: 24, md: 12 }}>
          <List
            itemLayout="horizontal"
            dataSource={backendTech}
            renderItem={(item) => (
              <List.Item>
                <List.Item.Meta
                  avatar={item.icon}
                  title={item.name}
                  description={item.desc}
                />
              </List.Item>
            )}
          />
        </ProCard>
      </ProCard.Group>
    </PageContainer>
  );
};

export default AboutPage;
