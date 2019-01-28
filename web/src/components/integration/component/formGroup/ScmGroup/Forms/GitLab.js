import React from 'react';
import { Field } from 'formik';
import { Input, Button } from 'antd';
import MakeField from '@/components/public/makeField';
import ValidateSelect from './ValidateSelect';

const InputField = MakeField(Input);
export default class GitLab extends React.Component {
  render() {
    return (
      <div>
        <Field
          label={intl.get('integration.form.scm.serverAddress')}
          name="spec.inline.scm.server"
          component={InputField}
        />
        <Field
          label={intl.get('integration.form.scm.verificationMode')}
          name="validateFunc"
          component={ValidateSelect}
        />
        <Button>验证</Button>
      </div>
    );
  }
}
