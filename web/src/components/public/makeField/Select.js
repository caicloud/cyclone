import PropTypes from 'prop-types';
import { Select } from 'antd';

const Option = Select.Option;
const SelectComponent = props => {
  const { value, disabled, handleSelectChange, options } = props;
  return (
    <Select
      value={value}
      disabled={disabled}
      placeholder={intl.get('integration.form.datasourcetype')}
      onChange={handleSelectChange}
    >
      {options.map(o => (
        <Option value={o.value} key={o.value}>
          {o.name}
        </Option>
      ))}
    </Select>
  );
};
SelectComponent.propTypes = {
  handleSelectChange: PropTypes.func,
  value: PropTypes.string,
  disabled: PropTypes.bool,
  options: ProcessingInstruction.array,
};

export default SelectComponent;
