import {
  getDisplayChangelog,
  getDisplayDescription,
  getDisplayName,
  isEnglishLocale,
  isPublicPluginRoute,
  pickLocaleText,
} from './plugin';

describe('plugin utils', () => {
  it('recognizes public plugin routes', () => {
    expect(isPublicPluginRoute('/plugins')).toBe(true);
    expect(isPublicPluginRoute('/plugins/12')).toBe(true);
    expect(isPublicPluginRoute('#/plugins')).toBe(true);
    expect(isPublicPluginRoute('/plugins/#/plugins/12')).toBe(true);
    expect(isPublicPluginRoute('/plugin/project-management')).toBe(false);
  });

  it('detects english locale and falls back between zh/en fields', () => {
    expect(isEnglishLocale('en-US')).toBe(true);
    expect(isEnglishLocale('zh-CN')).toBe(false);
    expect(pickLocaleText('en-US', '中文', 'English')).toBe('English');
    expect(pickLocaleText('en-US', '中文', '')).toBe('中文');
    expect(pickLocaleText('zh-CN', '????', 'English Name')).toBe('English Name');
    expect(getDisplayName('en-US', { nameZh: '插件中心', nameEn: '' })).toBe('插件中心');
    expect(getDisplayDescription('zh-CN', { descriptionZh: '', descriptionEn: 'Description' })).toBe(
      'Description',
    );
    expect(getDisplayChangelog('en-US', { changelogZh: '修复问题', changelogEn: '' })).toBe('修复问题');
  });
});
