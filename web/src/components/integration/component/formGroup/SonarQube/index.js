import React from 'react';
import { Field } from 'formik';
import PropTypes from 'prop-types';
import MakeField from '@/components/public/makeField';
import { Input } from 'antd';

const InputField = MakeField(Input);

export default class SonarQube extends React.Component {
  static propTypes = {
    setFieldValue: PropTypes.func,
  };
  changeConfig = value => {
    const { setFieldValue } = this.props;
    setFieldValue('types', value);
  };
  render() {
    return (
      <div>
        <Field
          label="Server地址"
          required
          name="spec.sonarQube.server"
          component={InputField}
          onChange={this.changeConfig}
        />
        <Field
          label="Token"
          required
          name="spec.sonarQube.token"
          component={InputField}
          onChange={this.changeConfig}
        />
      </div>
    );
  }
}
