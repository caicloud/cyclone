import React from 'react';
import { Field } from 'formik';
import Selection from './Selection';

export default class ScmGroup extends React.Component {
  render() {
    return (
      <div>
        <h3>代码源</h3>
        <Field
          label="类型"
          name="codeType"
          {...this.props}
          component={Selection}
        />
      </div>
    );
  }
}
