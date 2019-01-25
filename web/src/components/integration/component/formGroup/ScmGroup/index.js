import React from 'react';
import { Field } from 'formik';
import Selection from './Selection';
import PropTypes from 'prop-types';

export default class ScmGroup extends React.Component {
  static propTypes = {
    setFieldValue: PropTypes.func,
  };
  changeConfig = value => {
    const { setFieldValue } = this.props;
    setFieldValue('types', value);
  };
  render() {
    return (
      <Field
        label="类型"
        name="types"
        component={Selection}
        onChange={this.changeConfig}
      />
    );
  }
}
