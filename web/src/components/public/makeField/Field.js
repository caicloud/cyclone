import React from 'react';
import { Form, Icon, Tooltip } from 'antd';
import PropTypes from 'prop-types';
import { defaultFormItemLayout, noLabelItemLayout } from '@/lib/const';

const FormItem = Form.Item;

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
      tooltip: PropTypes.string,
    };
    const {
      label,
      hasFeedback,
      field,
      required = false,
      form: { touched, errors },
      formItemLayout,
      tooltip,
      ...rest
    } = props;
    const name = field.name;
    const hasError = _.get(touched, name) && _.get(errors, name);
    const _formItemLayout =
      formItemLayout || (label ? defaultFormItemLayout : noLabelItemLayout);
    const _label = tooltip ? (
      <span>
        {label}
        <Tooltip title={tooltip} placement="right">
          <Icon type="question-circle" style={{ marginLeft: '4px' }} />
        </Tooltip>
      </span>
    ) : (
      label
    );
    return (
      <FormItem
        label={_label}
        validateStatus={hasError ? 'error' : 'success'}
        hasFeedback={hasFeedback && hasError}
        help={hasError}
        required={required}
        {..._formItemLayout}
      >
        <Component {...field} {...rest} />
      </FormItem>
    );
  };
}
