import { Form, Input } from 'antd';
import { Field, FieldArray } from 'formik';
import MakeField from '@/components/public/makeField';
import { noLabelItemLayout } from '@/lib/const';
import PropTypes from 'prop-types';
import { inject, observer } from 'mobx-react';
import { modalFormItemLayout } from '@/lib/const';
import { required } from '@/components/public/validate';
import SelectPlus from '@/components/public/makeField/select';

const FormItem = Form.Item;
const InputField = MakeField(Input);
const SelectField = MakeField(SelectPlus);

@inject('resource')
@observer
class SCM extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      integrationName: '',
    };
  }
  render() {
    const {
      values,
      integrationList,
      setFieldValue,
      resource: { SCMRepos, listSCMRepos },
    } = this.props;
    const { integrationName } = this.state;
    const repoList = _.get(SCMRepos, `${integrationName}.items`, []);
    return (
      <FormItem {...noLabelItemLayout}>
        <FieldArray
          name="spec.parameters"
          render={() => (
            <div>
              {_.get(values, 'spec.parameters', []).map((field, index) => {
                if (field.name === 'SCM_TOKEN') {
                  return (
                    <Field
                      key={field.name}
                      label={intl.get('sideNav.integration')}
                      name={`spec.parameters.${index}.value`}
                      payload={{
                        items: integrationList,
                        nameKey: 'metadata.name',
                        valueKey: 'spec.scm.token',
                      }}
                      handleSelectChange={val => {
                        // note: make sour the token is unique
                        const item = _.find(
                          integrationList,
                          o => _.get(o, 'spec.scm.token') === val
                        );
                        const name = _.get(item, 'metadata.name');
                        if (!_.get(SCMRepos, name)) {
                          listSCMRepos(name);
                        }
                        this.setState({ integrationName: name });
                        setFieldValue(`spec.parameters.${index}.value`, val);
                      }}
                      component={SelectField}
                      required
                      validate={required}
                      formItemLayout={modalFormItemLayout}
                    />
                  );
                }
                if (field.name === 'SCM_URL') {
                  return (
                    <Field
                      key={field.name}
                      label="URL"
                      name={`spec.parameters.${index}.value`}
                      payload={{
                        items: repoList,
                        nameKey: 'name',
                        valueKey: 'url',
                      }}
                      handleSelectChange={val => {
                        setFieldValue(`spec.parameters.${index}.value`, val);
                      }}
                      component={SelectField}
                      formItemLayout={modalFormItemLayout}
                      required
                      validate={required}
                    />
                  );
                }
                return (
                  <Field
                    key={field.name}
                    label={
                      field.name.includes('SCM_')
                        ? field.name.replace('SCM_', '')
                        : field.name
                    }
                    name={`spec.parameters.${index}.value`}
                    component={InputField}
                    formItemLayout={modalFormItemLayout}
                    hasFeedback
                    required
                    validate={required}
                  />
                );
              })}
            </div>
          )}
        />
      </FormItem>
    );
  }
}

SCM.propTypes = {
  values: PropTypes.object,
  integrationList: PropTypes.array,
  setFieldValue: PropTypes.func,
  resource: PropTypes.object,
};

export default SCM;
