import React from 'react';
import { Field } from 'formik';
import { Input } from 'antd';
import MakeField from '@/components/public/makeField';

const InputField = MakeField(Input);
export default class SVN extends React.Component {
  render() {
    return (
      <div>
        <Field
          label="服务地址"
          name="spec.inline.scm.server"
          component={InputField}
          required
        />
        <Field
          label="用户名"
          name="spec.inline.scm.user"
          required
          component={InputField}
        />
        <Field
          label="密码"
          name="spec.inline.scm.password"
          required
          component={InputField}
        />
      </div>
    );
  }
}
