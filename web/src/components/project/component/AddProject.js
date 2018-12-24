import React from 'react';
import PropTypes from 'prop-types';
import { Modal, Form, Input, Select } from 'antd';

const FormItem = Form.Item;
const Option = Select.Option;
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
        >
          <Form layout="horizontal">
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
            <FormItem label={intl.get('project.authentication')}>
              {getFieldDecorator('authentication')(
                <Select>
                  <Option value="gitlab">gitlab</Option>
                  <Option value="github">github</Option>
                  <Option value="registry">registry</Option>
                </Select>
              )}
            </FormItem>
          </Form>
        </Modal>
      );
    }
  }
);

export default CreateProjectForm;
