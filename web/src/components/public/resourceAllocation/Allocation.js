import React from 'react';
import PropTypes from 'prop-types';
import { Radio, Input, Select } from 'antd';
import InputWithUnit from '@/public/inputWithUnit';

const RadioButton = Radio.Button;
const RadioGroup = Radio.Group;
const Option = Select.Option;

const allocationMap = [
  {
    name: 'basic',
    value: {
      cpuRequest: '0.5Core',
      cpuLimit: '1Core',
      memoryRequest: '1GiB',
      memoryLimit: '2GiB',
    },
  },
  {
    name: 'middle',
    value: {
      cpuRequest: '1Core',
      cpuLimit: '2Core',
      memoryRequest: '2 GiB',
      memoryLimit: '4 GiB',
    },
  },
  {
    name: 'high',
    value: {
      cpuRequest: '2Core',
      cpuLimit: '4Core',
      memoryRequest: '4GiB',
      memoryLimit: '8GiB',
    },
  },
];

class Allocation extends React.Component {
  state = {
    mode: 'recommend',
  };

  handleTypeSelect = e => {
    this.setState({ mode: e.target.value });
  };

  handleAllocationSelect = e => {
    const { onChange } = this.props;
    const value = e.target.value;
    const item = _.find(allocationMap, n => _.get(n, 'name') === value);
    onChange(_.get(item, 'value'));
  };

  handleCustomInput = (key, val) => {
    const { onChange, value } = this.props;
    onChange(_.merge({}, value, { [key]: val }));
  };

  render() {
    const { mode } = this.state;
    return (
      <div className="u-resource-allocation">
        <div className="allocation-type">
          <RadioGroup defaultValue="recommend" onChange={this.handleTypeSelect}>
            <RadioButton value="recommend">
              {intl.get('allocation.recommend')}
            </RadioButton>
            <RadioButton value="custom">
              {intl.get('allocation.custom')}
            </RadioButton>
          </RadioGroup>
        </div>
        <div className="allocation-content">
          {mode === 'recommend' ? (
            <RadioGroup onChange={this.handleAllocationSelect}>
              {allocationMap.map(a => (
                <RadioButton value={a.name} key={a.name}>
                  <div>
                    <div className="content">
                      <div>
                        <span className="key">
                          {intl.get('allocation.cpuRequest')}
                        </span>
                        <span>{parseFloat(a.value.cpuRequest)} Core</span>
                      </div>
                      <div>
                        <span className="key">
                          {intl.get('allocation.cpuLimit')}
                        </span>
                        <span>{parseFloat(a.value.cpuLimit)} Core</span>
                      </div>
                      <div>
                        <span className="key">
                          {intl.get('allocation.memoryRequest')}
                        </span>
                        <span>{parseFloat(a.value.memoryRequest)} GiB</span>
                      </div>
                      <div>
                        <span className="key">
                          {intl.get('allocation.memoryLimit')}
                        </span>
                        <span>{parseFloat(a.value.memoryLimit)} GiB</span>
                      </div>
                    </div>
                    <div className="footer">
                      {intl.get(`allocation.${a.name}`)}
                    </div>
                  </div>
                </RadioButton>
              ))}
            </RadioGroup>
          ) : (
            <div className="custom">
              <div>
                <div>
                  <span className="key">
                    {intl.get('allocation.cpuRequest')}:
                  </span>
                  <InputWithUnit
                    inputKey="cpuRequest"
                    addonAfter="Core"
                    defaultAddon="Core"
                    onChange={this.handleCustomInput}
                  />
                </div>
                <div>
                  <span className="key">
                    {intl.get('allocation.cpuLimit')}:
                  </span>
                  <InputWithUnit
                    inputKey="cpuLimit"
                    addonAfter="Core"
                    defaultAddon="Core"
                    onChange={this.handleCustomInput}
                  />
                </div>
              </div>
              <div>
                <div>
                  <span className="key">
                    {intl.get('allocation.memoryRequest')}:
                  </span>
                  <InputWithUnit
                    defaultAddon="MiB"
                    inputKey="memoryRequest"
                    addonAfter={[
                      { name: 'MiB', value: 'MiB' },
                      { name: 'GiB', value: 'GiB' },
                      { name: 'TiB', value: 'TiB' },
                    ]}
                    onChange={this.handleCustomInput}
                  />
                </div>
                <div>
                  <span className="key">
                    {intl.get('allocation.memoryLimit')}:
                  </span>
                  <InputWithUnit
                    defaultAddon="MiB"
                    inputKey="memoryLimit"
                    addonAfter={[
                      { name: 'MiB', value: 'MiB' },
                      { name: 'GiB', value: 'GiB' },
                      { name: 'TiB', value: 'TiB' },
                    ]}
                    onChange={this.handleCustomInput}
                  />
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    );
  }
}

export default Allocation;
