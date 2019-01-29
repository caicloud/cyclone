import React from 'react';
import { Field } from 'formik';
import Selection from './Selection';

export default class ScmGroup extends React.Component {
  render() {
    return (
      <div className="u-scm">
        <h3>代码源</h3>
        <Field
          label={intl.get('integration.type')}
          name="spec.scm.type"
          {...this.props}
          component={Selection}
        />
      </div>
    );
  }
}
