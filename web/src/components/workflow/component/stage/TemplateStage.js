import { Input, Form } from 'antd';
import { Field, FieldArray } from 'formik';
import SectionCard from '@/components/public/sectionCard';
import MakeField from '@/components/public/makeField';
import ResourceArray from '../resource/ResourceArray';
import PropTypes from 'prop-types';
import { required } from '@/components/public/validate';
import { defaultFormItemLayout } from '@/lib/const';

const Fragment = React.Fragment;
const { TextArea } = Input;
const FormItem = Form.Item;
const InputField = MakeField(Input);
const TextareaField = MakeField(TextArea);

class TemplateStage extends React.Component {
  static propTypes = {
    stageId: PropTypes.string,
    values: PropTypes.object,
    update: PropTypes.bool,
    project: PropTypes.string,
    modify: PropTypes.bool,
  };

  renderSection = (data, key) => {
    const dom = [];
    _.forEach(data.value, (v, k) => {
      dom.push(
        <Field
          key={key}
          label={k}
          name={`${key}.${k}`}
          component={InputField}
          hasFeedback
          required
          validate={required}
        />
      );
    });
    return dom;
  };
  render() {
    const { stageId, values, update, project, modify } = this.props;
    const specKey = `${stageId}.spec.pod`;
    return (
      <Fragment>
        {update && modify ? (
          <FormItem label={intl.get('name')} {...defaultFormItemLayout}>
            {_.get(values, `${stageId}.metadata.name`)}
          </FormItem>
        ) : (
          <Field
            label={intl.get('name')}
            name={`${stageId}.metadata.name`}
            component={InputField}
            hasFeedback
            required
            validate={required}
          />
        )}
        <SectionCard title={intl.get('input')}>
          <ResourceArray
            resourcesField={`${specKey}.inputs.resources`}
            resources={_.get(values, `${specKey}.inputs.resources`, [])}
            update={update}
            project={project}
          />
          <FieldArray
            name={`${specKey}.inputs.arguments`}
            render={arrayHelpers => (
              <div>
                {_.get(values, `${specKey}.inputs.arguments`, []).map(
                  (r, i) => {
                    if (_.isObject(r.value)) {
                      return this.renderSection(
                        r,
                        `${specKey}.inputs.arguments.${i}.value`
                      );
                    } else {
                      return (
                        <Field
                          key={i}
                          label={r.name}
                          name={`${specKey}.inputs.arguments.${i}.value`}
                          component={
                            ['cmd'].includes(r.name)
                              ? TextareaField
                              : InputField
                          }
                          hasFeedback
                          required
                          validate={required}
                        />
                      );
                    }
                  }
                )}
              </div>
            )}
          />
        </SectionCard>
      </Fragment>
    );
  }
}

export default TemplateStage;
