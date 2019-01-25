import React from 'react';
import { Field } from 'formik';
import { Input, Button } from 'antd';
import MakeField from '@/components/public/makeField';
import ValidateSelect from './ValidateSelect';

const InputField = MakeField(Input);
export default class GitHub extends React.Component {
  render() {
    return (
      <div>
        <Field
          label="服务地址"
          name="serviceAddress"
          component={InputField}
          onChange={this.changeConfig}
        />
        <Field
          label="验证方式"
          name="validateFunc"
          component={ValidateSelect}
          onChange={this.changeConfig}
        />
        <Button>验证</Button>
      </div>
    );
  }
}
