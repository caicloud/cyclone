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
          label={intl.get('integration.form.scm.serverAddress')}
          name="spec.inline.scm.server"
          component={InputField}
          required
        />
        <Field
          label={intl.get('integration.form.username')}
          name="spec.inline.scm.user"
          required
          component={InputField}
        />
        <Field
          label={intl.get('integration.form.pwd')}
          name="spec.inline.scm.password"
          required
          component={InputField}
        />
      </div>
    );
  }
}
