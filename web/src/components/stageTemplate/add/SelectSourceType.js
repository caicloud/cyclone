import PropTypes from 'prop-types';
import { Select } from 'antd';

const Option = Select.Option;
const SelectSourceType = props => {
  const value =
    props.value && props.value.length > 0 ? props.value[0].type : '';
  return (
    <Select
      value={value}
      disabled={props.disabled}
      placeholder={intl.get('template.form.newResourceType.placeholder')}
      onChange={props.handleSelectChange}
    >
      <Option value="SCM">SCM</Option>
      <Option value="Image">Image</Option>
    </Select>
  );
};
SelectSourceType.propTypes = {
  handleSelectChange: PropTypes.func,
  value: PropTypes.array,
  disabled: PropTypes.bool,
};

export default SelectSourceType;
