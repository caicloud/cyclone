import React from 'react';
import PropTypes from 'prop-types';
import { Form, Select, Input, Button } from 'antd';
import { Field, Formik } from 'formik';
import ScmGroup from '../formGroup/ScmGroup/index';
import SonarQube from '../formGroup/SonarQube/index';
import { inject, observer } from 'mobx-react';
import DockerRegistry from '../formGroup/DockerRegistry/index';
import MakeField from '@/components/public/makeField';
import integration from '../../../../store/integration';

const FormItem = Form.Item;
const Option = Select.Option;
const { TextArea } = Input;
const InputField = MakeField(Input);
const TextareaField = MakeField(TextArea);

const selectSourceType = props => {
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

selectSourceType.propTypes = {
  handleSelectChange: PropTypes.func,
};

const SelectField = MakeField(selectSourceType);

const generateData = (data, type) => {
  const metadata = {
    creationTime: Date.now(),
    description: '',
    name: '',
  };
  const spec = {
    inline: {
      dockerRegistry: {},
      general: [
        {
          name: '',
          value: '',
        },
      ],
      scm: {
        password: '',
        server: '',
        token: '',
        type: '',
        user: '',
      },
      sonarQube: {},
    },
    type: '',
  };
  _.forEach(metadata, (v, k) => {
    if (k in data) {
      v = data[k];
    }
  });
  _.forEach(spec.inline[type], (v, k) => {
    if (k in data) {
      v = data[k];
    }
  });
  const submitData = Object.assign({}, { metadata, spec });
  return submitData;
};

@inject('integration')
@observer
export default class IntegrationForm extends React.Component {
  static propTypes = {
    payload: PropTypes.object,
    history: PropTypes.object,
    errors: PropTypes.object,
    values: PropTypes.object,
    handleSubmit: PropTypes.func,
    setFieldValue: PropTypes.func,
    sourceType: PropTypes.string,
    initialFormData: PropTypes.object,
  };

  renderWrapForm = (sourceType, props) => {
    const formMap = {
      scm: <ScmGroup {...props} />,
      dockerRegistry: <DockerRegistry />,
      sonarQube: <SonarQube />,
    };
    return formMap[sourceType];
  };

  handleSelectChange = val => {
    const { setFieldValue } = this.props;
    setFieldValue('sourceType', val);
  };

  handleCancle = () => {
    const { history } = this.props;
    history.push('/integration');
  };

  render() {
    return (
      <Formik
        initialValues={{
          metadata: { name: '', description: '', creationTime: '' },
          sourceType: '',
          spec: {
            inline: {
              dockerRegistry: {
                password: '',
                server: '',
                user: '',
              },
              general: [
                {
                  name: '',
                  value: '',
                },
              ],
              scm: {
                password: '',
                server: 'https://github.com',
                token: '',
                type: 'GitHub',
                user: '',
              },
              sonarQube: {
                token: '',
                server: '',
              },
            },
          },
        }}
        validate={values => {
          const errors = {};
          const spec = {
            inline: {
              scm: {},
              sonarQube: {},
              dockerRegistry: {},
            },
          };
          if (!values.metadata.name) {
            errors.metadata = { name: '不能为空' };
          }
          if (!values.sourceType) {
            errors.sourceType = '请选择类型s';
          }

          if (!values.spec.inline.scm.server) {
            spec.inline.scm.server = '请填写服务地址';
            errors['spec'] = spec;
          }

          if (!values.spec.inline.scm.token) {
            spec.inline.scm.token = '请填写token信息';
            errors['spec'] = spec;
          }

          if (!values.spec.inline.scm.user) {
            spec.inline.scm.user = '请填写用户名';
            errors['spec'] = spec;
          }
          if (!values.spec.inline.scm.password) {
            errors.spec.inline.scm.password = '请填写密码';
          }
          return errors;
        }}
        onSubmit={values => {
          const { sourceType } = values;
          const dsubmitData = generateData(values, sourceType);
          integration.createIntegration(dsubmitData, () => {
            this.props.history.replace(`/integration`);
          });
        }}
        render={props => {
          const {
            handleSubmit,
            setFieldValue,
            errors,
            values: { sourceType },
          } = props;
          return (
            <Form onSubmit={handleSubmit}>
              <Field
                label={intl.get('integration.name')}
                name="metadata.name"
                component={InputField}
                hasFeedback
                required
              />
              <Field
                label={intl.get('integration.desc')}
                name="metadata.description"
                component={TextareaField}
              />
              <Field
                label={intl.get('integration.type')}
                name="sourceType"
                required
                handleSelectChange={val => {
                  setFieldValue('sourceType', val);
                }}
                component={SelectField}
              />
              {sourceType && this.renderWrapForm(sourceType, props)}
              <FormItem>
                <Button
                  style={{ marginRight: '10px' }}
                  onClick={this.handleCancle}
                >
                  {intl.get('integration.form.cancel')}
                </Button>
                <Button type="primary" htmlType="submit">
                  {intl.get('integration.form.confirm')}
                </Button>
                {!_.isEmpty(errors) && <p>有错误存在</p>}
              </FormItem>
            </Form>
          );
        }}
      />
    );
  }
}
