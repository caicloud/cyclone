import React from 'react';
import { Form } from 'antd';
import PropTypes from 'prop-types';

const FormItem = Form.Item;

const defaultFormItemLayout = {
  labelCol: { span: 4 },
  wrapperCol: { span: 14 },
};

export default function makeField(Component) {
  return function FieldWithProps(props) {
    FieldWithProps.propTypes = {
      label: PropTypes.string,
      hasFeedback: PropTypes.bool,
      field: PropTypes.object,
      form: PropTypes.object,
      children: PropTypes.node,
      required: PropTypes.bool,
      formItemLayout: PropTypes.object,
    };
    const {
      label,
      hasFeedback,
      field,
      required = false,
      form: { isValid, touched, errors },
      formItemLayout,
      ...rest
    } = props;
    const name = field.name;
    const hasError = touched[name] && !isValid;
    const _formItemLayout = formItemLayout || defaultFormItemLayout;
    return (
      <FormItem
        label={label}
        validateStatus={hasError ? 'error' : 'success'}
        hasFeedback={hasFeedback && hasError}
        help={hasError && errors[name]}
        required={required}
        {..._formItemLayout}
      >
        <Component {...field} {...rest} />
      </FormItem>
    );
  };
}
