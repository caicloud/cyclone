import React from 'react';
import { Select, Row, Col } from 'antd';
import { Form, Button } from 'antd';
import PropTypes from 'prop-types';
import { inject, observer } from 'mobx-react';
import { IntegrationTypeMap } from '@/consts/const.js';

const FormItem = Form.Item;
const { Option, OptGroup } = Select;

@inject('integration')
@observer
class IntegrationSelect extends React.Component {
  static propTypes = {
    label: PropTypes.string,
    hasFeedback: PropTypes.bool,
    field: PropTypes.object,
    form: PropTypes.object,
    children: PropTypes.node,
    integration: PropTypes.object,
  };
  constructor(props) {
    super(props);
    // TODO(qme): initialize when update
    this.state = {
      selectedSourceType: [],
    };
  }
  componentDidMount() {
    const { integration } = this.props;
    integration.getIntegrationList();
  }
  render() {
    const {
      label,
      hasFeedback,
      field,
      form: { isValid, touched, errors, setFieldValue },
      integration,
    } = this.props;
    const { selectedSourceType } = this.state;
    const list = integration.groupIntegrationList;
    const groupKeys = integration.getGroupKeys();
    const name = field.name;
    const value = field.value || [];
    const hasError = touched[name] && !isValid;
    const handleChange = value => {
      setFieldValue(name, value);
      const selected = _.map(value, n => n.split('/')[0]);
      this.setState({ selectedSourceType: selected });
    };
    const handleRemove = text => {
      const val = _.pull(value, text);
      setFieldValue(name, val);
      const selected = _.pull(selectedSourceType, text.split('/')[0]);
      this.setState({ selectedSourceType: selected });
    };
    return (
      <FormItem
        label={label}
        validateStatus={hasError ? 'error' : 'success'}
        hasFeedback={hasFeedback && hasError}
        help={hasError && errors[name]}
        {...{
          labelCol: { span: 4 },
          wrapperCol: { span: 14 },
        }}
      >
        <Select
          mode="multiple"
          value={value}
          name={name}
          onChange={handleChange}
        >
          {groupKeys.map(o => {
            return (
              <OptGroup label={o} key={o}>
                {list[o].map(v => (
                  <Option
                    value={`${o}/${v.metadata.name}`}
                    key={v.metadata.name}
                    disabled={
                      _.includes(selectedSourceType, o) &&
                      !_.includes(value, `${o}/${v.metadata.name}`)
                    }
                  >
                    {_.get(v, 'metadata.name')}
                  </Option>
                ))}
              </OptGroup>
            );
          })}
        </Select>
        {value.map(o => {
          const item = integration.getGroupItem(o);
          return (
            <div className="integration-item" key={o}>
              <Row>
                <Col span={12}>
                  <div className="text-item">
                    <div className="key">{intl.get('integration.name')}：</div>
                    <div className="value">{_.get(item, 'metadata.name')}</div>
                  </div>
                </Col>
                <Col span={12}>
                  <div className="text-item">
                    <div className="key">{intl.get('integration.type')}：</div>
                    <div className="value">{_.get(item, 'spec.type')}</div>
                  </div>
                </Col>
              </Row>
              <Row>
                <Col span={12}>
                  <div className="text-item">
                    <div className="key">
                      {intl.get('integration.serviceAddress')}：
                    </div>
                    <div className="value">
                      {_.get(
                        item,
                        `spec.${
                          IntegrationTypeMap[_.get(item, 'spec.type')]
                        }.server`
                      )}
                    </div>
                  </div>
                </Col>
              </Row>
              <Button
                type="dashed"
                shape="circle"
                icon="close"
                onClick={() => handleRemove(o)}
              />
            </div>
          );
        })}
      </FormItem>
    );
  }
}

export default IntegrationSelect;
