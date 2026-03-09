import { join } from 'node:path';
import { defineConfig } from '@umijs/max';
import defaultSettings from './defaultSettings';

import routes from './routes';

const { REACT_APP_ENV = 'dev', NODE_ENV } = process.env;

/**
 * @name 使用公共路径
 * @description 部署时的路径，如果部署在非根目录下，需要配置这个变量
 * @doc https://umijs.org/docs/api/config#publicpath
 */
const PUBLIC_PATH: string = '/';

export default defineConfig({
  // 浏览器 兼容性设置
  targets: {ie: 11},
  // 配置网站的标题
  title: defaultSettings.title,
  layout: {
    locale: true,
    ...defaultSettings,
  },
  // 路由的配置
  routes,
  // 配置路由模式
  history: {type: "hash"},
  // 开发环境生成 sourcemap，生产环境不生成
  devtool: NODE_ENV === 'development' ? 'eval' : false,
  // 问号传参时，可以通过query传参
  historyWithQuery: {},
  // 配置图片的打包方式，大于10KB单独打包成一个图片，反之打包成base64
  inlineLimit: 10000,
  // 配置js的压缩方式
  jsMinifier: 'terser',
  jsMinifierOptions: {},
  // 打包后的文件会带 hash 处理浏览器缓存问题
  hash: true,
  // 配置打包后资源的导入路径
  publicPath: PUBLIC_PATH,

  // 主题的配置；虽然叫主题，但是其实只是 less 的变量设置
  // theme: { '@primary-color': '#1DA57A' }

  // 国际化配置
  ignoreMomentLocale: true,

  // 代理配置，可以让你的本地服务器代理到你的服务器上，这样你就可以访问服务器的数据了；要注意以下 代理只能在本地开发时使用，build 之后就无法使用了。
  proxy: {
    '/api/': {
      target: 'http://127.0.0.1:8080',
      changeOrigin: true,
      pathRewrite: { '^api': 'api' },
    },
  },
  // 快速热更新配置
  fastRefresh: true,


  //============== 以下都是max的插件配置 ===============
  // 数据流插件
  model: {},
  // 一个全局的初始数据流，可以用它在插件之间共享数据；可以用来存放一些全局的数据，比如用户信息，或者一些全局的状态，全局初始状态在整个 Umi 项目的最开始创建。
  initialState: {},
  /**
   * @name layout 插件
   * @doc https://umijs.org/docs/max/layout-menu
   */

  /**
   * @name moment2dayjs 插件
   * @description 将项目中的 moment 替换为 dayjs
   * @doc https://umijs.org/docs/max/moment2dayjs
   */
  moment2dayjs: {
    preset: 'antd',
    plugins: ['duration'],
  },
  /**
   * @name 国际化插件
   * @doc https://umijs.org/docs/max/i18n
   */
  locale: {
    // default zh-CN
    default: 'zh-CN',
    antd: true,
    // default true, when it is true, will use `navigator.language` overwrite default
    baseNavigator: true,
  },
  /**
   * @name antd 插件
   * @description 内置了 babel import 插件
   * @doc https://umijs.org/docs/max/antd#antd
   */
  antd: {
    appConfig: {},
    configProvider: {
      theme: {
        cssVar: true,
        token: {
          fontFamily: 'AlibabaSans, sans-serif',
        },
      },
    },
  },
  /**
   * @name 网络请求配置
   * @description 它基于 axios 和 ahooks 的 useRequest 提供了一套统一的网络请求和错误处理方案。
   * @doc https://umijs.org/docs/max/request
   */
  request: {},
  /**
   * @name 权限插件
   * @description 基于 initialState 的权限插件，必须先打开 initialState
   * @doc https://umijs.org/docs/max/access
   */
  access: {},
  /**
   * @name <head> 中额外的 script
   * @description 配置 <head> 中额外的 script
   */
  headScripts: [
    // 解决首次加载时白屏的问题
    { src: join(PUBLIC_PATH, 'scripts/loading.js'), async: true },
  ],

  //================ pro 插件配置 =================
  presets: ['umi-presets-pro'],
  /**
   * @name openAPI 插件的配置
   * @description 基于 openapi 的规范生成serve 和mock，能减少很多样板代码
   * @doc https://pro.ant.design/zh-cn/docs/openapi/
   */
  openAPI: [
    {
      requestLibPath: "import { request } from '@umijs/max'",
      // 或者使用在线的版本
      // schemaPath: "https://gw.alipayobjects.com/os/antfincdn/M%24jrzTTYJN/oneapi.json"
      schemaPath: join(__dirname, 'oneapi.json'),
      mock: false,
    },
    {
      requestLibPath: "import { request } from '@umijs/max'",
      schemaPath: 'https://gw.alipayobjects.com/os/antfincdn/CA1dOm%2631B/openapi.json',
      projectName: 'swagger',
    },
  ],
  mock: {
    include: ['mock/**/*', 'src/pages/**/_mock.ts'],
  },
  /**
   * @name 是否开启 mako
   * @description 使用 mako 极速研发
   * @doc https://umijs.org/docs/api/config#mako
   */
  mako: {},
  esbuildMinifyIIFE: true,
  requestRecord: {},
  exportStatic: {},
  define: {
    'process.env.CI': process.env.CI,
  },
});
