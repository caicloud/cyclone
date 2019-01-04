import React from 'react';
import PropTypes from 'prop-types';
import { Form, Select, Input, Button } from 'antd';
import TextArea from 'antd/lib/input/TextArea';
import ScmGroup from './ScmGroup';

const FormItem = Form.Item;
const Option = Select.Option;

class DataForm extends React.Component {
  static propTypes = {
    form: PropTypes.object,
    onSubmit: PropTypes.func,
    onCancel: PropTypes.func,
  };
  constructor(props) {
    super(props);
    this.formMap = {
      SCM: <ScmGroup form={this.props.form} />,
      dockerregistry: null,
    };
    this.state = {
      subForm: null,
    };
  }

  componentDidMount() {
    this.resetForm();
  }

  componentWillUnmount() {
    this.resetForm();
  }

  resetForm = () => {
    const { resetFields } = this.props.form;
    this.setState({ subForm: null });
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
      subForm: this.formMap[value],
    });
  };

  render() {
    const { getFieldDecorator } = this.props.form;
    const { subForm } = this.state;
    return (
      <Form onSubmit={this.handleSubmit}>
        <FormItem
          label={intl.get('type')}
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
              <Option value="SCM">SCM</Option>
              <Option value="dockerregistry">
                {intl.get('integration.dataform.dockerregistry')}
              </Option>
            </Select>
          )}
        </FormItem>
        {subForm && subForm}
        <Form.Item wrapperCol={{ span: 8, offset: 10 }}>
          <Button style={{ marginRight: '10px' }} onClick={this.handleCancle}>
            {intl.get('integration.dataform.cancel')}
          </Button>
          <Button type="primary" htmlType="submit">
            {intl.get('integration.dataform.confirm')}
          </Button>
        </Form.Item>
      </Form>
    );
  }
}

export default Form.create()(DataForm);
