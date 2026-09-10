import { t, useI18n } from '@/i18n';
// src/page/about/index.tsx
import React from 'react';
import { PageContainer, ProCard } from '@ant-design/pro-components';
import { Descriptions, Tag, List, Typography, Space } from 'antd';
import {
  HeartOutlined,
  GithubOutlined,
  BookOutlined,
  ReadOutlined,
  CodeOutlined,
} from '@ant-design/icons';

// 假设的项目信息，你可以替换为你自己的

const AboutPage: React.FC = () => {
  useI18n();
  const projectInfo = {
    name: 'Base Frame CMS',
    version: 'base-frame',
    description: t('cms.aModernContentManagementSystemBuilt'),
    documentation: 'https://github.com/CIPFZ/go-web-frame/tree/base-frame/docs',
    // 替换为你的文档链接
    github: 'https://github.com/CIPFZ/go-web-frame/tree/base-frame', // 替换为你的仓库链接
  };

  // 前端技术栈
  const frontendTech = [
    {
      name: 'React',
      desc: t('cms.coreUiLibrary'),
      icon: (
        <ReadOutlined
          style={{
            color: '#61DAFB',
          }}
        />
      ),
    },
    {
      name: 'Ant Design Pro',
      desc: t('cms.enterpriseUiFramework'),
      icon: (
        <HeartOutlined
          style={{
            color: '#1890ff',
          }}
        />
      ),
    },
    {
      name: 'UmiJS',
      desc: t('cms.routingBuildAndDevelopmentTools'),
      icon: (
        <BookOutlined
          style={{
            color: '#1890ff',
          }}
        />
      ),
    },
    {
      name: 'TypeScript',
      desc: t('cms.typedJavascriptLanguage'),
      icon: (
        <CodeOutlined
          style={{
            color: '#3178C6',
          }}
        />
      ),
    },
  ];

  // 后端技术栈
  const backendTech = [
    {
      name: 'Go (Golang)',
      desc: t('cms.apiImplementationLanguage'),
      icon: (
        <CodeOutlined
          style={{
            color: '#00ADD8',
          }}
        />
      ),
    },
    {
      name: 'Gin',
      desc: t('cms.highPerformanceHttpFramework'),
      icon: (
        <CodeOutlined
          style={{
            color: '#00ADD8',
          }}
        />
      ),
    },
    {
      name: 'GORM',
      desc: t('cms.ormLibraryForGo'),
      icon: (
        <CodeOutlined
          style={{
            color: '#00ADD8',
          }}
        />
      ),
    },
    {
      name: 'MySQL/PostgreSQL',
      desc: t('cms.relationalDatabases'),
      icon: (
        <CodeOutlined
          style={{
            color: '#00ADD8',
          }}
        />
      ),
    },
  ];
  return (
    <PageContainer title={t('cms.about')}>
      <ProCard
        style={{
          marginBottom: 16,
        }}
        title={
          <Space>
            <GithubOutlined />
            <Typography.Title
              level={4}
              style={{
                margin: 0,
              }}
            >
              {projectInfo.name}
            </Typography.Title>
          </Space>
        }
      >
        <Typography.Paragraph>{projectInfo.description}</Typography.Paragraph>
        <Descriptions bordered size="small">
          <Descriptions.Item label={t('cms.currentVersion')}>
            <Tag color="blue">{projectInfo.version}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label={t('cms.documentation')}>
            <a
              href={projectInfo.documentation}
              target="_blank"
              rel="noopener noreferrer"
            >
              {t('cms.viewDocumentation')}
            </a>
          </Descriptions.Item>
          <Descriptions.Item label="Github">
            <a
              href={projectInfo.github}
              target="_blank"
              rel="noopener noreferrer"
            >
              {t('cms.sourceRepository')}
            </a>
          </Descriptions.Item>
        </Descriptions>
      </ProCard>

      <ProCard.Group title={t('cms.technologyStack')} direction="column">
        <ProCard
          title={t('cms.frontend')}
          colSpan={{
            xs: 24,
            md: 12,
          }}
        >
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

        <ProCard
          title={t('cms.backend')}
          colSpan={{
            xs: 24,
            md: 12,
          }}
        >
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
