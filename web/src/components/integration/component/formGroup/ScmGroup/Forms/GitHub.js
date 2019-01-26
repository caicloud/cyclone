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
          name="spec.inline.scm.server"
          disabled
          component={InputField}
        />
        <Field
          label="验证方式"
          name="validateFunc"
          component={ValidateSelect}
        />
        <Button>校验</Button>
      </div>
    );
  }
}
