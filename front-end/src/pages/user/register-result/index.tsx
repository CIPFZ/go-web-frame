import { t, useI18n } from '@/i18n';
import { Link, useSearchParams } from '@umijs/max';
import { Button, Result } from 'antd';
import { SelectLang } from '@/components';

export default function RegisterResult() {
  useI18n();
  const [params] = useSearchParams();
  const account = params.get('account');
  return (
    <>
      <div style={{ textAlign: 'right', padding: 16 }}>
        <SelectLang />
      </div>
      <Result
        status="success"
        title={
          account
            ? t('cms.registeredAccount', { account })
            : t('cms.accountCreated')
        }
        subTitle={t('cms.yourAccountIsReadyYouCan')}
        extra={
          <Link to="/user/login">
            <Button type="primary">{t('cms.backToSignIn')}</Button>
          </Link>
        }
      />
    </>
  );
}
