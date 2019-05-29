import PropTypes from 'prop-types';
import { AutoComplete } from 'antd';

const { Option } = AutoComplete;
const AutoCompleteField = props => {
  const {
    value,
    handleSelectChange,
    placeholder,
    payload: { items, nameKey = 'name', valueKey = 'value' },
  } = props;

  return (
    <AutoComplete
      onChange={handleSelectChange}
      placeholder={placeholder}
      value={value}
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
    </AutoComplete>
  );
};

AutoCompleteField.propTypes = {
  handleSelectChange: PropTypes.func,
  value: PropTypes.string,
  placeholder: PropTypes.string,
  payload: PropTypes.shape({
    items: PropTypes.array,
    nameKey: PropTypes.string,
    valueKey: PropTypes.string,
  }),
};

export default AutoCompleteField;
