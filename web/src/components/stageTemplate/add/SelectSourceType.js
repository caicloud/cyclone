import PropTypes from 'prop-types';
import { Select } from 'antd';

const Option = Select.Option;
const SelectSourceType = props => {
  const value = props.value && !_.isEmpty(props.value) ? props.value.type : '';
  return (
    <Select
      value={value}
      disabled={props.disabled}
      placeholder={intl.get('template.form.newResourceType.placeholder')}
      onChange={props.handleSelectChange}
    >
      <Option value="Git">Git</Option>
      <Option value="SVN">SVN</Option>
      <Option value="Image">Image</Option>
    </Select>
  );
};
SelectSourceType.propTypes = {
  handleSelectChange: PropTypes.func,
  value: PropTypes.object,
  disabled: PropTypes.bool,
};

export default SelectSourceType;
