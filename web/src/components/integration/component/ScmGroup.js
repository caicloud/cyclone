import React from 'react';
import { Input, Radio, Form } from 'antd';
import PropTypes from 'prop-types';
const RadioGroup = Radio.Group;
const RadioButton = Radio.Button;
const FormItem = Form.Item;

export default class ScmGroup extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      subForm: null,
      curCvs: null,
      validateType: 'Token',
    };
    this.formMap = {
      GitHub: this.renderSubGitForm(),
      GitLab: this.renderSubGitForm(),
      SVN: this.renderSubSVNForm(),
    };
  }

  componentDidMount() {
    this.setState({
      subForm: this.formMap['GitHub'],
      curCvs: 'GitHub',
    });
  }

  handleSwitchCvs = e => {
    const value = _.get(e, 'target.value');
    this.setState({
      subForm: this.formMap[value],
      curCvs: value,
    });
  };

  handleSwitchValidate = e => {
    const value = _.get(e, 'target.value');
    this.setState({
      validateType: value,
    });
  };

  renderSubGitForm = () => {
    const { getFieldDecorator } = this.props.form;
    const { validateType } = this.state;
    return (
      <div>
        <FormItem
          label="验证方式"
          labelCol={{ span: 5 }}
          wrapperCol={{ span: 12 }}
        >
          {getFieldDecorator('validateType', {
            rules: [
              {
                required: true,
                message: intl.get('integration.dataform.datasourcetype'),
              },
            ],
            initialValue: 'Token',
          })(
            <RadioGroup onChange={this.handleSwitchValidate}>
              <RadioButton value="Token">Token</RadioButton>
              <RadioButton value="UsPwd">用户名和密码</RadioButton>
            </RadioGroup>
          )}
        </FormItem>
        {this.renderValidateForm(validateType)}
      </div>
    );
  };

  renderSubSVNForm = () => {
    return <div>adfsd</div>;
  };

  renderValidateForm = type => {
    switch (type) {
      case 'Token':
        return <div>Token</div>;
      case 'UsPwd':
        return <div>UsPwd</div>;
      default:
        return null;
    }
  };

  render() {
    const { getFieldDecorator } = this.props.form;
    const { subForm, curCvs } = this.state;
    return (
      <div>
        <p>代码源</p>
        <FormItem
          label={intl.get('integration.type')}
          labelCol={{ span: 5 }}
          wrapperCol={{ span: 12 }}
        >
          {getFieldDecorator('cvsType', {
            rules: [
              {
                required: true,
                message: intl.get('integration.dataform.datasourcetype'),
              },
            ],
            initialValue: 'GitHub',
          })(
            <RadioGroup onChange={this.handleSwitchCvs}>
              <RadioButton value="GitHub">GitHub</RadioButton>
              <RadioButton value="GitLab">GitLab</RadioButton>
              <RadioButton value="SVN">SVN</RadioButton>
            </RadioGroup>
          )}
        </FormItem>
        <FormItem
          label="服务地址"
          labelCol={{ span: 5 }}
          wrapperCol={{ span: 12 }}
        >
          {curCvs !== 'GitHub' ? (
            getFieldDecorator('serviceAddress', {
              rules: [
                {
                  required: true,
                  message: intl.get('integration.dataform.datasourcetype'),
                },
              ],
            })(<Input />)
          ) : (
            <p>https://github.com</p>
          )}
        </FormItem>
        {subForm && subForm}
      </div>
    );
  }
}
ScmGroup.propTypes = {
  form: PropTypes.object.isRequired,
};
