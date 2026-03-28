import React from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import { Empty, Typography } from 'antd';

const ReviewWorkbenchPage: React.FC = () => {
  return (
    <PageContainer title="审核工作台">
      <Empty
        description={
          <Typography.Text type="secondary">
            Transitional placeholder. The real review workbench will be implemented in a later task.
          </Typography.Text>
        }
      />
    </PageContainer>
  );
};

export default ReviewWorkbenchPage;
