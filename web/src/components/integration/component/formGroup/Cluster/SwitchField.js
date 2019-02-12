import React from 'react';
import { Form, Switch } from 'antd';
import PropTypes from 'prop-types';

const FormItem = Form.Item;

const defaultFormItemLayout = {
  labelCol: { span: 4 },
  wrapperCol: { span: 16 },
};

const FieldSwitch = props => {
  const {
    label,
    field: { value },
    formItemLayout,
    onChange,
  } = props;
  const _formItemLayout = formItemLayout || defaultFormItemLayout;
  return (
    <FormItem label={label} {..._formItemLayout}>
      <Switch defaultChecked={value} onChange={onChange} />
    </FormItem>
  );
};

FieldSwitch.propTypes = {
  label: PropTypes.string,
  field: PropTypes.object,
  formItemLayout: PropTypes.object,
  onChange: PropTypes.func,
};

export default FieldSwitch;
