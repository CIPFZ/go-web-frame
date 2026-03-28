import React from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import { Empty, Typography } from 'antd';

const PublishWorkbenchPage: React.FC = () => {
  return (
    <PageContainer title="发布工作台">
      <Empty
        description={
          <Typography.Text type="secondary">
            Transitional placeholder. The real publish workbench will be implemented in a later task.
          </Typography.Text>
        }
      />
    </PageContainer>
  );
};

export default PublishWorkbenchPage;
