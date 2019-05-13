import { Input } from 'antd';
import { Field, FieldArray } from 'formik';
import SectionCard from '@/components/public/sectionCard';
import MakeField from '@/components/public/makeField';
import ResourceArray from '../resource/ResourceArray';
import PropTypes from 'prop-types';

const Fragment = React.Fragment;
const { TextArea } = Input;
const InputField = MakeField(Input);
const TextareaField = MakeField(TextArea);

class TemplateStage extends React.Component {
  static propTypes = {
    stageId: PropTypes.string,
    values: PropTypes.object,
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
        />
      );
    });
    return dom;
  };
  render() {
    const { stageId, values } = this.props;
    return (
      <Fragment>
        <Field
          label={intl.get('name')}
          name={`${stageId}.name`}
          component={InputField}
          hasFeedback
          required
        />
        <SectionCard title={intl.get('input')}>
          <ResourceArray
            resourcesField={`${stageId}.spec.pod.inputs.resources`}
            resources={_.get(
              values,
              `${stageId}.spec.pod.inputs.resources`,
              []
            )}
          />
          <FieldArray
            name={`${stageId}.spec.pod.inputs.arguments`}
            render={arrayHelpers => (
              <div>
                {_.get(values, `${stageId}.spec.pod.inputs.arguments`, []).map(
                  (r, i) => {
                    if (_.isObject(r.value)) {
                      return this.renderSection(
                        r,
                        `${stageId}.spec.pod.inputs.arguments.${i}.value`
                      );
                    } else {
                      return (
                        <Field
                          key={i}
                          label={r.name}
                          name={`${stageId}.spec.pod.inputs.arguments.${i}.value`}
                          component={
                            ['cmd'].includes(r.name)
                              ? TextareaField
                              : InputField
                          }
                          hasFeedback
                          required
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
