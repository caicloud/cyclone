import PropTypes from 'prop-types';
import { Select } from 'antd';

const Option = Select.Option;
const SelectSourceType = props => {
  return (
    <Select
      defaultValue={props.value}
      placeholder={intl.get('integration.form.datasourcetype')}
      onChange={props.handleSelectChange}
    >
      <Option value="scm">SCM</Option>
      <Option value="DockerRegistry">
        {intl.get('integration.form.dockerregistry')}
      </Option>
      <Option value="SonarQube">SonarQube</Option>
    </Select>
  );
};
SelectSourceType.propTypes = {
  handleSelectChange: PropTypes.func,
  value: PropTypes.string,
};

export default SelectSourceType;
