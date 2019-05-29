import PropTypes from 'prop-types';
import { Select } from 'antd';

const Option = Select.Option;
const SelectSourceType = props => {
  const value = props.value && !_.isEmpty(props.value) ? props.value.type : '';
  const { options = [] } = props;
  return (
    <Select
      value={value}
      disabled={props.disabled}
      placeholder={intl.get('template.form.newResourceType.placeholder')}
      onChange={props.handleSelectChange}
    >
      {options.length > 0 &&
        options.map((o, i) => (
          <Option key={i} value={o}>
            {o}
          </Option>
        ))}
    </Select>
  );
};
SelectSourceType.propTypes = {
  handleSelectChange: PropTypes.func,
  value: PropTypes.object,
  disabled: PropTypes.bool,
  options: PropTypes.array,
};

export default SelectSourceType;
