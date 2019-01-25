import React from 'react';
import { Field } from 'formik';
import PropTypes from 'prop-types';
import MakeField from '@/components/public/makeField';
import { Input } from 'antd';

const InputField = MakeField(Input);

export default class DockerRegistry extends React.Component {
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
          label="仓库地址"
          required
          name="registryAddress"
          component={InputField}
          onChange={this.changeConfig}
        />
        <Field
          label="管理员账号"
          required
          name="adminUser"
          component={InputField}
          onChange={this.changeConfig}
        />
        <Field
          label="管理员密码"
          required
          name="adminPwd"
          component={InputField}
          onChange={this.changeConfig}
        />
      </div>
    );
  }
}
