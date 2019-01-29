import React from 'react';
import { Field } from 'formik';
import { Input } from 'antd';
import MakeField from '@/components/public/makeField';
import ValidateSelect from './ValidateSelect';

const InputField = MakeField(Input);
export default class GitHub extends React.Component {
  render() {
    return (
      <div>
        <Field
          label={intl.get('integration.form.scm.serverAddress')}
          name="spec.scm.server"
          disabled
          component={InputField}
        />
        <Field
          label={intl.get('integration.form.scm.verificationMode')}
          name="validateMode"
          component={ValidateSelect}
        />
      </div>
    );
  }
}
