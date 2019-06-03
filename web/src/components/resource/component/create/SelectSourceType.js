import PropTypes from 'prop-types';
import { Select } from 'antd';

const Option = Select.Option;
const SelectSourceType = props => {
  return (
    <Select
      value={props.value}
      disabled={props.disabled}
      placeholder={intl.get('integration.form.datasourcetype')}
      onChange={props.handleSelectChange}
    >
      <Option value="SCM">SCM</Option>
      <Option value="DockerRegistry">
        {intl.get('integration.form.dockerregistry')}
      </Option>
      <Option value="SonarQube">SonarQube</Option>
      <Option value="Cluster">
        {intl.get('integration.form.cluster.name')}
      </Option>
    </Select>
  );
};
SelectSourceType.propTypes = {
  handleSelectChange: PropTypes.func,
  value: PropTypes.string,
  disabled: PropTypes.bool,
};

export default SelectSourceType;
