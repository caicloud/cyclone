import PropTypes from 'prop-types';
import { Select } from 'antd';

const Option = Select.Option;
const SelectSourceType = props => {
  return (
    <Select
      placeholder={intl.get('integration.form.datasourcetype')}
      onChange={props.handleSelectChange}
    >
      <Option value="scm">SCM</Option>
      <Option value="dockerRegistry">
        {intl.get('integration.form.dockerregistry')}
      </Option>
      <Option value="sonarQube">SonarQube</Option>
    </Select>
  );
};
SelectSourceType.propTypes = {
  handleSelectChange: PropTypes.func,
};

export default SelectSourceType;
