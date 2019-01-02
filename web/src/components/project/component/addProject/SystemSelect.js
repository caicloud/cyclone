import React from 'react';
import { Select } from 'antd';
const { Option, OptGroup } = Select;

const SystemSelect = props => {
  return (
    <div>
      <Select mode="multiple">
        <OptGroup label="SCM">
          <Option value="scm/github">github111</Option>
          <Option value="scm/gitlab">gitlab222</Option>
        </OptGroup>
        <OptGroup label="Docker Registry">
          <Option value="docker/devops">devops</Option>
        </OptGroup>
        <OptGroup label="SonarQube">
          <Option value="sonarqube/test">test</Option>
        </OptGroup>
      </Select>
    </div>
  );
};

export default SystemSelect;
