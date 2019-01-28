import React from 'react';
import { Field } from 'formik';
import Selection from './Selection';

export default class ScmGroup extends React.Component {
  render() {
    return (
      <div>
        <h3>代码源</h3>
        <Field
          label={intl.get('integration.type')}
          name="spec.inline.scm.type"
          {...this.props}
          component={Selection}
        />
      </div>
    );
  }
}
