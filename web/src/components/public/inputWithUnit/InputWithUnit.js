import React from 'react';
import { Input, Select, Form } from 'antd';
import PropTypes from 'prop-types';
import classnames from 'classnames';

const FormItem = Form.Item;
const Option = Select.Option;

const memoryUnit = ['Mi', 'Gi', 'Ti'];
class InputAddon extends React.Component {
  static propTypes = {
    defaultAddon: PropTypes.string,
    addonAfter: PropTypes.oneOfType([PropTypes.string, PropTypes.node]),
    onChange: PropTypes.func,
    inputKey: PropTypes.string,
    form: PropTypes.object,
    field: PropTypes.object,
    HandleAddon: PropTypes.func,
    label: PropTypes.string,
    className: PropTypes.string,
    onBlur: PropTypes.func,
    type: PropTypes.oneOf(['cpu', 'memory']),
  };

  static defaultProps = {
    addonAfter: [
      { name: 'MiB', value: 'Mi' },
      { name: 'GiB', value: 'Gi' },
      { name: 'TiB', value: 'Ti' },
    ],
    defaultAddon: 'Mi',
  };

  constructor(props) {
    super(props);
    this.state = {
      byteFieldNum: props.field.value ? parseFloat(props.field.value) : '',
      byteUnit: props.type === 'cpu' ? '' : this.getUnit(props.field.value),
    };
  }

  getUnit = value => {
    let _unit = this.props.defaultAddon;
    const num = parseFloat(value);
    const unit = _.replace(value, num, '');
    if (_.includes(memoryUnit, unit)) {
      _unit = unit;
    }
    return _unit;
  };

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
    const {
      addonAfter,
      label,
      field,
      className,
      form: { errors, touched },
      type,
    } = this.props;
    const { byteFieldNum, byteUnit } = this.state;
    const cls = classnames('u-input-unit', { [className]: !!className });
    const name = field.name;
    const hasError = _.get(touched, name) && _.get(errors, name);
    const addonComp =
      type === 'memory' ? (
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
        'Core'
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
