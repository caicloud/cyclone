import React from 'react';
import { Input, Select } from 'antd';
import PropTypes from 'prop-types';

const Option = Select.Option;

class InputAddon extends React.Component {
  static propTypes = {
    defaultAddon: PropTypes.string,
    addonAfter: PropTypes.oneOfType([PropTypes.string, PropTypes.node]),
    onChange: PropTypes.func,
    inputKey: PropTypes.string,
  };
  constructor(props) {
    super(props);
    this.state = {
      addonAfterText: props.defaultAddon,
      inputValue: null,
    };
  }

  handleChange = e => {
    const { onChange, inputKey } = this.props;
    const { addonAfterText } = this.state;
    const value = e.target.value;
    this.setState({ inputValue: value });
    onChange(inputKey, `${value}${addonAfterText}`);
  };

  handleAddon = value => {
    const { onChange, inputKey } = this.props;
    const { inputValue } = this.state;
    this.setState({ addonAfterText: value });
    onChange(inputKey, `${inputValue}${value}`);
  };
  render() {
    const { addonAfter, defaultAddon } = this.props;
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

    return <Input addonAfter={addonComp} onChange={this.handleChange} />;
  }
}

export default InputAddon;
