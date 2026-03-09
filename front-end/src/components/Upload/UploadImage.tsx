import React, { useState } from 'react';
import { Upload, message } from 'antd';
import { LoadingOutlined, PlusOutlined } from '@ant-design/icons';
import type { UploadChangeParam } from 'antd/es/upload';
import type { RcFile, UploadFile, UploadProps } from 'antd/es/upload/interface';

// 获取 Token
const getToken = () => localStorage.getItem('token') || '';

interface UploadImageProps {
  value?: string;
  onChange?: (url: string) => void;
  disabled?: boolean;
  action?: string;
  data?: Record<string, unknown> | ((file: UploadFile) => Record<string, unknown>);

  // ✨✨✨ 新增属性：是否为圆形 ✨✨✨
  circle?: boolean;
}

const UploadImage: React.FC<UploadImageProps> = ({
                                                   value,
                                                   onChange,
                                                   disabled,
                                                   action = '/api/v1/sys/user/avatar',
                                                   data,
                                                   // ✨ 默认为 true，保证个人中心头像依然是圆的
                                                   circle = true,
                                                 }) => {
  const [loading, setLoading] = useState(false);

// 文件上传前的校验 (保持不变)
  const beforeUpload = (file: RcFile) => {
    const isJpgOrPng = file.type === 'image/jpeg' || file.type === 'image/png' || file.type === 'image/gif' || file.type === 'image/webp';
    if (!isJpgOrPng) {
      message.error('只能上传 JPG/PNG/GIF 文件!');
    }
    const isLt2M = file.size / 1024 / 1024 < 2;
    if (!isLt2M) {
      message.error('图片大小不能超过 2MB!');
    }
    return isJpgOrPng && isLt2M;
  };

  const handleChange: UploadProps['onChange'] = (info: UploadChangeParam<UploadFile>) => {
    if (info.file.status === 'uploading') {
      setLoading(true);
      return;
    }

    if (info.file.status === 'done') {
      setLoading(false);
      const response = info.file.response;

      // 检查 response 是否存在
      if (!response) {
        message.error('上传响应为空');
        return;
      }

      // 检查 code 是否为 0
      // 注意：有时候后端返回的是字符串 "0"，用 == 比较更稳妥，或者确认类型
      if (response.code === 0) {
        let url = response.data?.url; // 使用可选链防止 crash

        if (url) {
          // 追加时间戳防缓存
          const separator = url.includes('?') ? '&' : '?';
          url = `${url}${separator}t=${new Date().getTime()}`;

          if (onChange) {
            onChange(url);
          } else {
            console.warn('⚠️ [Upload Debug] onChange 未定义！组件可能未正确绑定 Form.Item');
          }
          message.success('上传成功');
        } else {
          console.error('❌ [Upload Debug] data.url 未找到');
        }
      } else {
        message.error(response?.msg || '上传失败');
      }
    } else if (info.file.status === 'error') {
      setLoading(false);
      message.error('上传网络错误');
    }
  };

  const uploadButton = (
    <div>
      {loading ? <LoadingOutlined /> : <PlusOutlined />}
      <div style={{ marginTop: 8 }}>上传</div>
    </div>
  );

  return (
    <Upload
      name="file"
      listType="picture-card"
      className="avatar-uploader"
      showUploadList={false}
      action={action}
      data={data}
      headers={{ 'x-token': getToken() }}
      beforeUpload={beforeUpload}
      onChange={handleChange}
      disabled={disabled}
    >
      {value ? (
        <img
          src={value}
          alt="avatar"
          style={{
            width: '100%',
            height: '100%',
            objectFit: 'cover',
            // ✨✨✨ 关键修改：根据 circle 属性决定圆角 ✨✨✨
            // true: 50% (圆形)
            // false: 8px (圆角矩形，更美观) 或 0 (直角)
            borderRadius: circle ? '50%' : '8px'
          }}
        />
      ) : (
        uploadButton
      )}
    </Upload>
  );
};

export default UploadImage;
