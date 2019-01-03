import React from 'react';
import PropTypes from 'prop-types';
import { Form, Select, Input, Button } from 'antd';

const FormItem = Form.Item;
const Option = Select.Option;
const formMap = {
  git: [
    {
      label: 'URL',
      key: 'url',
    },
    {
      label: 'Token',
      key: 'Token',
    },
  ],
  imagesregistry: [
    {
      label: intl.get('integration.dataform.registryaddress'),
      key: 'registryAddress',
    },
    {
      label: intl.get('integration.dataform.username'),
      key: 'username',
    },
    {
      label: intl.get('integration.dataform.pwd'),
      key: 'pwd',
    },
  ],
};

class DataForm extends React.Component {
  static propTypes = {
    form: PropTypes.object,
    onSubmit: PropTypes.func,
    onCancel: PropTypes.func,
  };
  state = {
    inputList: [],
  };

  componentDidMount() {
    this.resetForm();
  }

  componentWillUnmount() {
    this.resetForm();
  }

  resetForm = () => {
    const { resetFields } = this.props.form;
    this.setState({ inputList: [] });
    resetFields();
  };

  handleSubmit = e => {
    const { onSubmit } = this.props;
    e.preventDefault();
    this.props.form.validateFields((err, values) => {
      if (!err) {
        // TODO post request
        this.resetForm();
        onSubmit();
      }
    });
  };

  handleCancle = () => {
    const { onCancel } = this.props;
    this.resetForm();
    onCancel();
  };

  handleSelectChange = value => {
    this.setState({
      inputList: formMap[value],
    });
  };

  renderFormInputs = () => {
    const { inputList } = this.state;
    const { getFieldDecorator } = this.props.form;
    return (
      inputList.length > 0 &&
      inputList.map((v, i) => (
        <FormItem
          key={i}
          label={v.label}
          labelCol={{ span: 5 }}
          wrapperCol={{ span: 12 }}
        >
          {getFieldDecorator(v.key)(<Input autoComplete="off" />)}
        </FormItem>
      ))
    );
  };

  render() {
    const { getFieldDecorator } = this.props.form;
    const { inputList } = this.state;
    return (
      <Form onSubmit={this.handleSubmit}>
        <FormItem
          label={intl.get('integration.type')}
          labelCol={{ span: 5 }}
          wrapperCol={{ span: 12 }}
        >
          {getFieldDecorator('type', {
            rules: [
              {
                required: true,
                message: intl.get('integration.dataform.datasourcetype'),
              },
            ],
          })(
            <Select
              placeholder={intl.get('integration.dataform.datasourcetype')}
              onChange={this.handleSelectChange}
            >
              <Option value="git">git</Option>
              <Option value="imagesregistry">
                {intl.get('integration.dataform.imagesregistry')}
              </Option>
            </Select>
          )}
        </FormItem>
        {this.renderFormInputs()}
        <Form.Item wrapperCol={{ span: 8, offset: 10 }}>
          <Button style={{ marginRight: '10px' }} onClick={this.handleCancle}>
            {intl.get('integration.dataform.cancel')}
          </Button>
          <Button
            disabled={inputList.length <= 0}
            type="primary"
            htmlType="submit"
          >
            {intl.get('integration.dataform.confirm')}
          </Button>
        </Form.Item>
      </Form>
    );
  }
}

export default Form.create()(DataForm);
