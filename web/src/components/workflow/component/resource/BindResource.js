import { Form, Input, Radio } from 'antd';
import { Field, FieldArray } from 'formik';
import MakeField from '@/components/public/makeField';
import { noLabelItemLayout, modalFormItemLayout } from '@/lib/const';
import PropTypes from 'prop-types';
import { inject, observer } from 'mobx-react';
import { required } from '@/components/public/validate';
import SelectPlus from '@/components/public/makeField/select';
import SCM from './SCM';

const RadioButton = Radio.Button;
const RadioGroup = Radio.Group;
const FormItem = Form.Item;
const InputField = MakeField(Input);
const SelectField = MakeField(SelectPlus);

const Fragment = React.Fragment;

// use in add stage, select a exist resource or create a new resource
@inject('integration')
@observer
class BindResource extends React.Component {
  static propTypes = {
    setFieldValue: PropTypes.func,
    values: PropTypes.object,
    type: PropTypes.oneOf(['inputs', 'outputs']),
    integration: PropTypes.object,
    update: PropTypes.boolean,
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
    const { values, setFieldValue, type, integration, update } = this.props;
    const resourceType = _.get(values, 'type', 'Git');
    const resourceList = _.get(integration, `groupIntegrationList.SCM`);
    const inputArray =
      type === 'inputs'
        ? [{ name: 'Git', value: 'Git' }]
        : [{ name: 'Image', value: 'Image' }];
    return (
      <Form layout={'horizontal'}>
        {/* <FormItem
          label={intl.get('workflow.resourceType')}
          {...modalFormItemLayout}
        >
          {type === 'inputs' ? 'Git' : 'Image'}
        </FormItem>
        {/* // TODO(qme): Subsequent support for multiple resource types */}
        <Field
          label={intl.get('workflow.resourceType')}
          name="type"
          handleSelectChange={val => {
            setFieldValue('type', val);
          }}
          payload={{
            items: inputArray,
          }}
          component={SelectField}
          formItemLayout={modalFormItemLayout}
          required
          validate={required}
        />
        {type === 'inputs' && (
          <FormItem
            label={intl.get('workflow.addMethod')}
            {...modalFormItemLayout}
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
            handleSelectChange={val => {
              setFieldValue('name', val);
            }}
            component={<div>TODO: resource select</div>}
            formItemLayout={modalFormItemLayout}
            required
            validate={required}
          />
        ) : (
          <Fragment>
            {update ? (
              <FormItem label={intl.get('name')} {...modalFormItemLayout}>
                {_.get(values, 'name')}
              </FormItem>
            ) : (
              <Field
                label={intl.get('name')}
                name="name"
                component={InputField}
                formItemLayout={modalFormItemLayout}
                hasFeedback
                required
                validate={required}
              />
            )}
            {resourceType === 'Git' ? (
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
                            formItemLayout={modalFormItemLayout}
                            hasFeedback
                            required
                            validate={required}
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
        {type === 'inputs' && (
          <Field
            label={intl.get('workflow.usePath')}
            name="path"
            component={InputField}
            hasFeedback
            required
            validate={required}
            formItemLayout={modalFormItemLayout}
          />
        )}
      </Form>
    );
  }
}

export default BindResource;
