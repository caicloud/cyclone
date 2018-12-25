import React from 'react';
import PropTypes from 'prop-types';
import { Modal, Form, Input, Cascader } from 'antd';

const FormItem = Form.Item;
const { TextArea } = Input;
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
                <Cascader
                  options={[
                    {
                      label: 'SCM',
                      value: 'scm',
                      children: [
                        { value: 'githuab', label: 'Github' },
                        { value: 'gitlab', label: 'Gitlab' },
                      ],
                    },
                    {
                      label: 'Docker Registry',
                      value: 'registry',
                      children: [{ value: 'devops', label: 'Devops' }],
                    },
                  ]}
                />
              )}
            </FormItem>
            {/* TODO(qme)：resource component */}
          </Form>
        </Modal>
      );
    }
  }
);

export default CreateProjectForm;
