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
    form: PropTypes.object,
    field: PropTypes.object,
    HandleAddon: PropTypes.func,
    label: PropTypes.string,
    className: PropTypes.string,
  };

  constructor(props) {
    super(props);
    this.state = {
      addonAfterText: props.defaultAddon,
      inputValue: null,
    };
  }

  handleAddon = value => {
    const { HandleAddon } = this.props;
    this.setState({ addonAfterText: value });
    HandleAddon && HandleAddon(value);
  };

  render() {
    const {
      addonAfter,
      defaultAddon,
      label,
      form: { isValid, touched, errors },
      field,
      className,
    } = this.props;
    const cls = classnames('u-input-unit', { [className]: !!className });
    const name = field.name;
    const hasError = _.get(touched, name) && !isValid;
    const addonComp = _.isArray(addonAfter) ? (
      <Select
        defaultValue={defaultAddon}
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

    return (
      <div className={cls}>
        <FormItem
          label={label}
          validateStatus={hasError ? 'error' : 'success'}
          hasFeedback={hasError}
          help={hasError && _.get(errors, name)}
          required
          {...{ labelCol: { span: 8 }, wrapperCol: { span: 16 } }}
        >
          <Input {...field} addonAfter={addonComp} />
        </FormItem>
      </div>
    );
  }
}

export default InputAddon;
