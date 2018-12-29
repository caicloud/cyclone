import React from 'react';
import PropTypes from 'prop-types';
import { Modal, Form, Input, Select } from 'antd';
import ResourceAllocation from '@/public/resourceAllocation';

const FormItem = Form.Item;
const { TextArea } = Input;
const { Option, OptGroup } = Select;
/**
 * TODO: list authentication
 */
const CreateProjectForm = Form.create()(
  // eslint-disable-next-line
  class extends React.Component {
    static propTypes = {
      visible: PropTypes.bool,
      onCancel: PropTypes.func,
      onCreate: PropTypes.func,
      form: PropTypes.object,
    };

    checkResource = (rule, value, callback) => {
      console.log('rule', rule, value);
      // callback('超出');
    };

    handleDelete = (value, option) => {
      console.log('value', value, option);
    };
    render() {
      const { visible, onCancel, onCreate, form } = this.props;
      const { getFieldDecorator } = form;
      return (
        <Modal
          visible={visible}
          title={intl.get('project.createProject')}
          okText={intl.get('operation.confirm')}
          onCancel={onCancel}
          onOk={onCreate}
          width={600}
        >
          <Form>
            <FormItem label={intl.get('name')}>
              {getFieldDecorator('name', {
                rules: [
                  {
                    required: true,
                    message: intl.get('project.formTip.projectNameRequired'),
                  },
                ],
              })(<Input />)}
            </FormItem>
            <FormItem label={intl.get('description')}>
              {getFieldDecorator('description')(
                <TextArea autosize={{ minRows: 2, maxRows: 6 }} />
              )}
            </FormItem>
            <FormItem label="外部系统">
              {getFieldDecorator('system', {
                rules: [
                  {
                    required: true,
                  },
                ],
              })(
                <Select mode="multiple" onDeselect={this.handleDelete}>
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
              )}
            </FormItem>
            <FormItem label="资源配置">
              {getFieldDecorator('resource', {
                rules: [{ validator: this.checkResource }],
              })(<ResourceAllocation />)}
            </FormItem>
          </Form>
        </Modal>
      );
    }
  }
);

export default CreateProjectForm;
