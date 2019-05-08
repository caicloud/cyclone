import { Form, Input, Radio } from 'antd';
import { Field, FieldArray } from 'formik';
import MakeField from '@/components/public/makeField';
import { defaultFormItemLayout, noLabelItemLayout } from '@/lib/const';
import PropTypes from 'prop-types';
import { inject, observer } from 'mobx-react';
import SelectPlus from '@/components/public/makeField/select';
import SCM from './SCM';

const RadioButton = Radio.Button;
const RadioGroup = Radio.Group;
const FormItem = Form.Item;
const InputField = MakeField(Input);
const SelectField = MakeField(SelectPlus);
const Fragment = React.Fragment;

const inputArray = ['SCM', 'DockerRegistry', 'Cluster', 'SonarQube'];

// use in add stage, select a exist resource or create a new resource
@inject('integration')
@observer
class BindResource extends React.Component {
  static propTypes = {
    setFieldValue: PropTypes.func,
    values: PropTypes.object,
    type: PropTypes.oneOf(['input', 'output']),
    integration: PropTypes.object,
  };

  state = { addWay: 'new' };

  componentDidMount() {
    const { integration } = this.props;
    integration.getIntegrationList();
  }

  changeAddWay = value => {
    this.setState({ addWay: value });
  };

  render() {
    const { addWay } = this.state;
    const { values, setFieldValue, type, integration } = this.props;
    const resourceType = _.get(values, 'resourceType', 'SCM');
    const resourceList = _.get(
      integration,
      `groupIntegrationList.${resourceType}`
    );
    // TODO(qme): i18n
    return (
      <Form layout={'horizontal'}>
        {type === 'input' ? (
          <Field
            label={intl.get('type')}
            name="resourceType"
            required
            handleSelectChange={val => {
              setFieldValue('resourceType', val);
              // TODO(qme): set value
              // setFieldValue('spec.parameters', resourceParametersField[val]);
            }}
            payload={{
              items: inputArray,
            }}
            component={SelectField}
          />
        ) : (
          <FormItem
            label={intl.get('workflow.resourceType')}
            {...defaultFormItemLayout}
          >
            image
          </FormItem>
        )}
        {type === 'input' && (
          <FormItem
            label={intl.get('workflow.addMethod')}
            {...defaultFormItemLayout}
          >
            <RadioGroup onChange={this.changeAddWay} defaultValue={addWay}>
              <RadioButton value="new">{intl.get('operation.add')}</RadioButton>
              <RadioButton value="exist">
                {intl.get('workflow.existResource')}
              </RadioButton>
            </RadioGroup>
          </FormItem>
        )}
        {addWay === 'exist' ? (
          <Field
            label={intl.get('type')}
            name="name"
            required
            handleSelectChange={val => {
              setFieldValue('name', val);
            }}
            component={<div>TODO: resource select</div>}
          />
        ) : (
          <Fragment>
            <Field
              label={intl.get('name')}
              name="name"
              component={InputField}
              hasFeedback
              required
            />
            {resourceType === 'SCM' ? (
              <SCM
                values={values}
                integrationList={resourceList}
                setFieldValue={setFieldValue}
              />
            ) : (
              <FormItem {...noLabelItemLayout}>
                <FieldArray
                  name="spec.parameters"
                  render={() => (
                    <div>
                      {_.get(values, 'spec.parameters', []).map(
                        (field, index) => (
                          <Field
                            key={field.name}
                            label={
                              field.name.includes('GIT_')
                                ? field.name.replace('GIT_', '')
                                : field.name
                            }
                            name={`spec.parameters.${index}.value`}
                            component={InputField}
                            hasFeedback
                            required
                          />
                        )
                      )}
                    </div>
                  )}
                />
              </FormItem>
            )}
          </Fragment>
        )}
        {type === 'input' && (
          <Field
            label={intl.get('workflow.usePath')}
            name="path"
            component={InputField}
            hasFeedback
            required
          />
        )}
      </Form>
    );
  }
}

export default BindResource;
