import type { ProFormInstance } from '@ant-design/pro-components';
import { useEffect, useRef } from 'react';
import { useI18n } from './index';

// Re-translate visible errors without resetting entered values or validating
// untouched fields when a user changes language inside an open form.
export function useFormLocale() {
  const locale = useI18n();
  const ref = useRef<ProFormInstance | undefined>(undefined);
  useEffect(() => {
    const form = ref.current;
    const names = form
      ?.getFieldsError()
      .filter((field) => field.errors.length > 0)
      .map((field) => field.name);
    if (form && names?.length) void form.validateFields(names).catch(() => {});
  }, [locale]);
  return ref;
}
