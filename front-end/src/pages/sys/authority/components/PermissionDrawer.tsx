import React, {useEffect, useState} from 'react';
import { HomeOutlined, HomeFilled } from '@ant-design/icons';
import {Drawer, Tree, Spin, Button, message, Space, Tabs, Checkbox, Tooltip} from 'antd';
import type {CheckboxChangeEvent} from 'antd/es/checkbox';
import {useRequest} from '@umijs/max';

// --- 导入 API ---
import {getMenuAuthority, getMenuList} from '@/services/api/menu';
// ✨ 修正：确保引用路径精确到文件 (除非你有 api.ts 导出)
import {getApiList} from '@/services/api/api';
import {getPolicyPathByAuthorityId, updateCasbin} from '@/services/api/casbin';
import {setAuthorityMenus, updateAuthority} from '@/services/api/authority';

import type {AuthorityItem} from '../index';

// --- 类型定义 ---
type TreeDataItem = {
  title: string;
  key: string | number;
  path?: string;
  children?: TreeDataItem[];
};

// --- 辅助函数 ---

const normalizeMenuTree = (menuList: any[]): TreeDataItem[] => {
  if (!menuList) return [];
  return menuList.map((menu) => ({
    title: `${menu.name} ${menu.name !== menu.path ? `(${menu.path})` : ''}`,
    key: menu.ID,
    path: menu.path,
    children: menu.routes ? normalizeMenuTree(menu.routes) : [],
  }));
};

const buildApiTree = (apiList: any[]): TreeDataItem[] => {
  const groups: Record<string, TreeDataItem> = {};
  apiList.forEach((api) => {
    if (!groups[api.apiGroup]) {
      groups[api.apiGroup] = {
        title: api.apiGroup,
        key: `group:${api.apiGroup}`,
        children: [],
      };
    }
    const apiKey = `${api.path}:${api.method}`;
    groups[api.apiGroup].children?.push({
      title: `${api.description} [${api.method}]`,
      key: apiKey,
    });
  });
  return Object.values(groups);
};

const getAllKeys = (tree: TreeDataItem[]): (string | number)[] => {
  let keys: (string | number)[] = [];
  for (const item of tree) {
    keys.push(item.key);
    if (item.children) {
      keys = keys.concat(getAllKeys(item.children));
    }
  }
  return keys;
};

// ---------------- 组件实现 ----------------

type PermissionDrawerProps = {
  open: boolean;
  role: AuthorityItem;
  onClose: () => void;
  // 成功回调
  onSuccess?: () => void;
};

const PermissionDrawer: React.FC<PermissionDrawerProps> = ({ open, role, onClose, onSuccess }) => {
  const [activeTab, setActiveTab] = useState<string>('menu');
  const [menuTree, setMenuTree] = useState<TreeDataItem[]>([]);
  const [menuCheckedKeys, setMenuCheckedKeys] = useState<React.Key[]>([]);
  const [allMenuKeys, setAllMenuKeys] = useState<React.Key[]>([]);
  const [apiTree, setApiTree] = useState<TreeDataItem[]>([]);
  const [apiCheckedKeys, setApiCheckedKeys] = useState<React.Key[]>([]);
  const [defaultRouter, setDefaultRouter] = useState<string>('');

  // --- 1. 加载所有菜单数据 ---
  const { loading: menuLoading } = useRequest(
    async () => getMenuList({ pageInfo: { page: 1, pageSize: 9999 } }),
    {
      onSuccess: (res: any) => {
        // ✨ 关键修复：先打印原始 res 确认结构
        console.log('Raw Menu Response:', res);
        const tree = normalizeMenuTree(res);
        console.log('Tree:', tree);
        setMenuTree(tree);
        setAllMenuKeys(getAllKeys(tree));
      },
    }
  );

  // --- 2. 加载所有 API 数据 ---
  const { loading: apiLoading } = useRequest(
    async () => getApiList({ page: 1, pageSize: 9999 }),
    {
      onSuccess: (res: any) => {
        console.log('Raw Menu Response:', res);
        const list = res?.list || [];
        setApiTree(buildApiTree(list));
      },
    }
  );

  // --- 3. 回显当前角色的权限 ---
  useEffect(() => {
    if (open && role) {
      setDefaultRouter(role.defaultRouter || 'dashboard/workplace');
      // 3a. 回显 API 权限
      getPolicyPathByAuthorityId({ authorityId: String(role.authorityId) }).then((res: any) => {
        const responseData = res.code !== undefined ? res : res.data;
        if (responseData?.code === 0) {
          const rawData = responseData.data;
          const list = Array.isArray(rawData) ? rawData : (rawData?.list || []);

          const keys = list.map((item: any) => `${item.path}:${item.method}`);
          setApiCheckedKeys(keys);
        }
      });

      // 3b. 回显菜单权限
      getMenuAuthority({ authorityId: role.authorityId }).then((res: any) => {
        const responseData = res.code !== undefined ? res : res.data;
        if (responseData?.code === 0) {
          const rawData = responseData.data;
          const list = Array.isArray(rawData) ? rawData : (rawData?.list || []);

          const checkedIds = list.map((item: any) => item.ID);
          setMenuCheckedKeys(checkedIds);
        }
      });
    }
  }, [open, role]);


  const onMenuCheck = (checked: any) => {
    setMenuCheckedKeys(checked.checked ? checked.checked : checked);
  };

  const onMenuSelectAll = (e: CheckboxChangeEvent) => {
    setMenuCheckedKeys(e.target.checked ? allMenuKeys : []);
  };

  const onApiCheck = (checked: any) => {
    setApiCheckedKeys(checked.checked ? checked.checked : checked);
  };

  const handleSave = async () => {
    try {
      if (activeTab === 'menu') {
        // 1. 保存菜单权限
        const p1 = setAuthorityMenus({
          authorityId: role.authorityId,
          menuIds: menuCheckedKeys.map((k) => Number(k)),
        });

        // 2. ✨ 保存默认路由 (同时必须传 name，防止后端置空)
        const p2 = updateAuthority({
          authorityId: role.authorityId,
          authorityName: role.authorityName,
          defaultRouter: defaultRouter,
        });

        await Promise.all([p1, p2]);
        message.success('菜单权限及首页配置保存成功');
        // ✨ 关键修改：如果存在 onSuccess 回调，则调用它
        if (onSuccess) {
          onSuccess();
        }
      } else if (activeTab === 'api') {
        const casbinInfos = apiCheckedKeys
          .map((k) => String(k))
          .filter((k) => k.includes(':') && !k.startsWith('group:'))
          .map((k) => {
            const [path, method] = k.split(':');
            return { path, method };
          });

        await updateCasbin({
          authorityId: String(role.authorityId),
          casbinInfos,
        });
        message.success('API 权限保存成功');
      }
    } catch (error) {
      console.error(error);
      message.error('保存失败，请重试');
    }
  };

  // ✨ 自定义树节点渲染 (显示小房子图标)
  const renderTreeTitle = (nodeData: any) => {
    const isHome = defaultRouter === nodeData.path;
    return (
      <div className="group" style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', width: '100%' }}>
        <span>{nodeData.title}</span>

        {/* 仅当节点有 path 时才显示设置首页图标 */}
        {nodeData.path && (
          <Tooltip title={isHome ? "当前首页" : "设为首页"}>
            <span
              onClick={(e) => {
                e.stopPropagation(); // 阻止触发勾选
                setDefaultRouter(nodeData.path);
              }}
              style={{ marginLeft: 12, cursor: 'pointer', display: 'flex', alignItems: 'center' }}
            >
              {isHome ? (
                <HomeFilled style={{ color: '#faad14', fontSize: 16 }} />
              ) : (
                <HomeOutlined style={{ color: '#d9d9d9', fontSize: 16 }} />
              )}
            </span>
          </Tooltip>
        )}
      </div>
    );
  };

  const isLoading = menuLoading || apiLoading;

  const tabItems = [
    {
      key: 'menu',
      label: '角色菜单',
      children: (
        <>
          <div style={{ marginBottom: 16, paddingBottom: 8, borderBottom: '1px solid #f0f0f0' }}>
            <Checkbox
              onChange={onMenuSelectAll}
              checked={
                allMenuKeys.length > 0 && menuCheckedKeys.length === allMenuKeys.length
              }
              indeterminate={
                menuCheckedKeys.length > 0 && menuCheckedKeys.length < allMenuKeys.length
              }
            >
              全选 / 全不选
            </Checkbox>
          </div>
          <Tree
            checkable
            checkStrictly
            defaultExpandAll
            // ✨ 显式指定字段，防止 title/key 不匹配
            fieldNames={{ title: 'title', key: 'key', children: 'children' }}
            treeData={menuTree}
            checkedKeys={menuCheckedKeys}
            onCheck={onMenuCheck}
            // ✨ 移除 height，使用样式控制
            style={{ maxHeight: '600px', overflowY: 'auto' }}
            titleRender={renderTreeTitle}
          />
        </>
      ),
    },
    {
      key: 'api',
      label: '角色 API',
      children: (
        <>
          <div style={{ marginBottom: 16, color: '#888', fontSize: '12px' }}>
            * 勾选对应的 API 分组或具体接口即可授权
          </div>
          <Tree
            checkable
            defaultExpandAll={false}
            treeData={apiTree}
            checkedKeys={apiCheckedKeys}
            onCheck={onApiCheck}
            // ✨ 移除 height，使用样式控制
            style={{ maxHeight: '600px', overflowY: 'auto' }}
          />
        </>
      ),
    },
  ];

  return (
    <Drawer
      title={`角色配置 - ${role.authorityName}`}
      width={600}
      open={open}
      onClose={onClose}
      maskClosable={false}
      footer={
        <Space style={{ float: 'right' }}>
          <Button onClick={onClose}>取消</Button>
          <Button type="primary" onClick={handleSave} loading={isLoading}>
            保存当前 Tab 配置
          </Button>
        </Space>
      }
    >
      <Spin spinning={isLoading}>
        <Tabs activeKey={activeTab} onChange={setActiveTab} items={tabItems} />
      </Spin>
    </Drawer>
  );
};

export default PermissionDrawer;
