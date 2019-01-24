import React from 'react';
import { Radio } from 'antd';
import FormItem from 'antd/lib/form/FormItem';

const RadioButton = Radio.Button;
export default class ScmGroup extends React.Component {
  render() {
    return (
      <FormItem>
        <RadioButton>GitHub</RadioButton>
        <RadioButton>GitLab</RadioButton>
        <RadioButton>SVN</RadioButton>
      </FormItem>
    );
  }
}
