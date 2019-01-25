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
          name="serviceAddress"
          component={InputField}
          required
          onChange={this.changeConfig}
        />
        <Field
          label="用户名"
          name="username"
          required
          component={InputField}
          onChange={this.changeConfig}
        />
        <Field
          label="密码"
          name="pwd"
          required
          component={InputField}
          onChange={this.changeConfig}
        />
      </div>
    );
  }
}
