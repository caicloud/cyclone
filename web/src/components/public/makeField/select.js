import PropTypes from 'prop-types';
import { Select } from 'antd';

const Option = Select.Option;
const SelectPlus = props => {
  const {
    value,
    disabled,
    handleSelectChange,
    labelInValue = false,
    payload: { items, nameKey = 'name', valueKey = 'value' },
  } = props;
  return (
    <Select
      value={value}
      disabled={disabled}
      labelInValue={labelInValue}
      onChange={handleSelectChange}
    >
      {items.map(o => {
        const name = _.isObject(o) && nameKey ? _.get(o, nameKey) : o;
        const value = _.isObject(o) && valueKey ? _.get(o, valueKey) : o;
        return (
          <Option value={value} key={value}>
            {name}
          </Option>
        );
      })}
    </Select>
  );
};
SelectPlus.propTypes = {
  handleSelectChange: PropTypes.func,
  value: PropTypes.string,
  disabled: PropTypes.bool,
  labelInValue: PropTypes.bool,
  payload: PropTypes.shape({
    items: PropTypes.array,
    nameKey: PropTypes.string,
    valueKey: PropTypes.string,
  }),
};

export default SelectPlus;
