import React from 'react';
import { Input, Select, Form } from 'antd';
import PropTypes from 'prop-types';
import classnames from 'classnames';

const FormItem = Form.Item;
const Option = Select.Option;

class InputAddon extends React.Component {
  static propTypes = {
    defaultAddon: PropTypes.string,
    addonAfter: PropTypes.oneOfType([PropTypes.string, PropTypes.node]),
    onChange: PropTypes.func,
    inputKey: PropTypes.string,
    error: PropTypes.string,
    touched: PropTypes.bool,
    field: PropTypes.object,
    HandleAddon: PropTypes.func,
    label: PropTypes.string,
    className: PropTypes.string,
    onBlur: PropTypes.func,
  };

  constructor(props) {
    super(props);
    this.state = {
      byteFieldNum: '',
      byteUnit: props.defaultAddon || 'Mi',
    };
  }

  handleAddon = value => {
    const { byteFieldNum } = this.state;

    this.setState({ byteUnit: value });
    const valueWithUnit = `${byteFieldNum}${value}`;
    this.onChange(valueWithUnit);
  };

  inputChange = e => {
    const { byteUnit } = this.state;
    const value = e.target.value;
    this.setState({ byteFieldNum: value });
    const valueWithUnit = `${value}${byteUnit}`;
    this.onChange(valueWithUnit);
  };

  inputBlur = e => {
    const { field, onBlur } = this.props;
    field && field.onBlur && field.onBlur(e);
    onBlur && onBlur(e);
  };

  onChange = value => {
    const numberValue = parseFloat(value);
    const realValue = _.isNaN(numberValue) ? '' : value;
    const { onChange, field } = this.props;
    onChange && onChange(field.name, realValue);
  };

  render() {
    const { addonAfter, label, error, touched, field, className } = this.props;
    const { byteFieldNum, byteUnit } = this.state;
    const cls = classnames('u-input-unit', { [className]: !!className });
    const hasError = touched && error;
    const addonComp = _.isArray(addonAfter) ? (
      <Select
        defaultValue={byteUnit}
        style={{ width: 80 }}
        onChange={this.handleAddon}
      >
        {addonAfter.map(o => (
          <Option value={o.value} key={o.value}>
            {o.name}
          </Option>
        ))}
      </Select>
    ) : (
      addonAfter
    );

    const _attr = { name: field.name };
    _attr.onChange = this.inputChange;
    _attr.value = byteFieldNum;
    _attr.onBlur = this.inputBlur;
    return (
      <div className={cls}>
        <FormItem
          label={label}
          validateStatus={hasError ? 'error' : 'success'}
          hasFeedback={hasError}
          help={hasError}
          required
          {...{ labelCol: { span: 8 }, wrapperCol: { span: 16 } }}
        >
          <Input {..._attr} addonAfter={addonComp} />
        </FormItem>
      </div>
    );
  }
}

export default InputAddon;
