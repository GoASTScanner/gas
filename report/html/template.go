// (c) Copyright 2016 Hewlett Packard Enterprise Development LP
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package html

const templateContent = `
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>Golang Security Checker</title>
  <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/bulma/0.9.2/css/bulma.min.css" integrity="sha512-byErQdWdTqREz6DLAA9pCnLbdoGGhXfU6gm1c8bkf7F51JVmUBlayGe2A31VpXWQP+eiJ3ilTAZHCR3vmMyybA==" crossorigin="anonymous"/>
  <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/10.7.2/styles/default.min.css" integrity="sha512-kZqGbhf9JTB4bVJ0G8HCkqmaPcRgo88F0dneK30yku5Y/dep7CZfCnNml2Je/sY4lBoqoksXz4PtVXS4GHSUzQ==" crossorigin="anonymous"/>
  <script type="text/javascript" src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/10.7.2/highlight.min.js" integrity="sha512-s+tOYYcC3Jybgr9mVsdAxsRYlGNq4mlAurOrfNuGMQ/SCofNPu92tjE7YRZCsdEtWL1yGkqk15fU/ark206YTg==" crossorigin="anonymous"></script>
  <script type="text/javascript" src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/10.7.2/languages/go.min.js" integrity="sha512-+UYV2NyyynWEQcZ4sMTKmeppyV331gqvMOGZ61/dqc89Tn1H40lF05ACd03RSD9EWwGutNwKj256mIR8waEJBQ==" crossorigin="anonymous"></script>
  <script type="text/javascript" src="https://cdnjs.cloudflare.com/ajax/libs/react/15.3.2/react.min.js" integrity="sha256-cLWs9L+cjZg8CjGHMpJqUgKKouPlmoMP/0wIdPtaPGs=" crossorigin="anonymous"></script>
  <script type="text/javascript" src="https://cdnjs.cloudflare.com/ajax/libs/react/15.3.2/react-dom.min.js" integrity="sha256-JIW8lNqN2EtqC6ggNZYnAdKMJXRQfkPMvdRt+b0/Jxc=" crossorigin="anonymous"></script>
  <script type="text/javascript" src="https://cdnjs.cloudflare.com/ajax/libs/babel-standalone/6.17.0/babel.min.js" integrity="sha256-1IWWLlCKFGFj/cjryvC7GDF5wRYnf9tSvNVVEj8Bm+o=" crossorigin="anonymous"></script>
  <style>
  .field-label {
    min-width: 85px;
  }
  .break-word {
    word-wrap: break-word;
  }
  .help {
    white-space: pre-wrap;
  }
  .tag {
    width: 80px;
  }
  .panel-block{
    padding: 0;
  }
  </style>
</head>
<body>
  <section class="section">
    <div class="container">
      <div id="content"></div>
    </div>
  </section>
  <script>
    var data = {{ . }};
  </script>
  <script type="text/babel">
    var IssueTag = React.createClass({
      render: function() {
        var level = "tag "
        if (this.props.level === "HIGH") {
          level += "is-danger";
        } else if (this.props.level === "MEDIUM") {
          level += "is-warning";
        } else if (this.props.level === "LOW") {
          level += "is-info";
        }
        return (
          <div className="control">
            <div className="tags has-addons">
              <span className="tag is-dark">{ this.props.label }</span>
              <span className={ level }>{ this.props.level }</span>
            </div>
          </div>
        );
      }
    });
    var Issue = React.createClass({
      render: function() {
        return (
          <div className="issue box">
          <div className="columns">
              <div className="column is-three-fifths">
                <strong className="break-word">{ this.props.data.file } (line { this.props.data.line })</strong>
                <p>{ this.props.data.details }</p>
              </div>
              <div className="column is-two-fifths">
                <div className="field is-grouped is-grouped-multiline">
                  <IssueTag label="Severity" level={ this.props.data.severity }/>
                  <IssueTag label="Confidence" level={ this.props.data.confidence }/>
                </div>
              </div>
            </div>
            <div className="highlight">
              <pre>
                <code className="go hljs">{ this.props.data.code }</code>
              </pre>
            </div>
          </div>
        );
      }
    });
    var Stats = React.createClass({
      render: function() {
        return (
          <p className="help is-pulled-right">
            Gosec {this.props.data.GosecVersion} scanned { this.props.data.Stats.files.toLocaleString() } files
            with { this.props.data.Stats.lines.toLocaleString() } lines of code.
            { this.props.data.Stats.nosec ? '\n' + this.props.data.Stats.nosec.toLocaleString() + ' false positives (nosec) have been waived.' : ''}
          </p>
        );
      }
    });
    var Issues = React.createClass({
      render: function() {
        if (this.props.data.Stats.files === 0) {
          return (
            <div className="notification">
              No source files found. Do you even Go?
            </div>
          );
        }
        if (this.props.data.Issues.length === 0) {
          return (
            <div>
              <div className="notification">
                Awesome! No issues found!
              </div>
              <Stats data={ this.props.data } />
            </div>
          );
        }
        var issues = this.props.data.Issues
          .filter(function(issue) {
            return this.props.severity.includes(issue.severity);
          }.bind(this))
          .filter(function(issue) {
            return this.props.confidence.includes(issue.confidence);
          }.bind(this))
          .filter(function(issue) {
            if (this.props.issueType) {
              return issue.details.toLowerCase().startsWith(this.props.issueType.toLowerCase());
            } else {
              return true
            }
          }.bind(this))
          .map(function(issue) {
            return (<Issue data={issue} />);
          }.bind(this));
        if (issues.length === 0) {
          return (
            <div>
              <div className="notification">
                No issues matched given filters
                (of total { this.props.data.Issues.length } issues).
              </div>
              <Stats data={ this.props.data } />
            </div>
          );
        }
        return (
          <div className="issues">
            { issues }
            <Stats data={ this.props.data } />
          </div>
        );
      }
    });
    var LevelSelector = React.createClass({
      handleChange: function(level) {
        return function(e) {
          var updated = this.props.selected
            .filter(function(item) { return item != level; });
          if (e.target.checked) {
            updated.push(level);
          }
          this.props.onChange(updated);
        }.bind(this);
      },
      render: function() {
        var HIGH = "HIGH", MEDIUM = "MEDIUM", LOW = "LOW";
        var highDisabled = !this.props.available.includes(HIGH);
        var mediumDisabled = !this.props.available.includes(MEDIUM);
        var lowDisabled = !this.props.available.includes(LOW);
        return (
          <div className="field is-grouped">
            <div className="control is-expanded">
            <label className="checkbox" disabled={ highDisabled }>
              <input
                type="checkbox"
                checked={ this.props.selected.includes(HIGH) }
                disabled={ highDisabled }
                onChange={ this.handleChange(HIGH) }/> High
            </label>
            </div>
            <div className="control is-expanded">
            <label className="checkbox" disabled={ mediumDisabled }>
              <input
                type="checkbox"
                checked={ this.props.selected.includes(MEDIUM) }
                disabled={ mediumDisabled }
                onChange={ this.handleChange(MEDIUM) }/> Medium
            </label>
            </div>
            <div className="control is-expanded">
            <label className="checkbox" disabled={ lowDisabled }>
              <input
                type="checkbox"
                checked={ this.props.selected.includes(LOW) }
                disabled={ lowDisabled }
                onChange={ this.handleChange(LOW) }/> Low
            </label>
            </div>
          </div>
        );
      }
    });
    var Navigation = React.createClass({
      updateSeverity: function(vals) {
        this.props.onSeverity(vals);
      },
      updateConfidence: function(vals) {
        this.props.onConfidence(vals);
      },
      updateIssueType: function(e) {
        if (e.target.value == "all") {
          this.props.onIssueType(null);
        } else {
          this.props.onIssueType(e.target.value);
        }
      },
      render: function() {
        var issueTypes = this.props.allIssueTypes
          .map(function(it) {
            var matches = this.props.issueType == it
            return (
              <option value={ it } selected={ matches }>
                { it }
              </option>
            );
          }.bind(this));
        return (
          <nav className="panel">
            <div className="panel-heading">Filters</div>
            <div className="panel-block">
            <div className="box">
            <div className="field is-horizontal">
              <div className="field-label is-normal">
                <label className="label">Severity</label>
              </div>
              <div className="field-body">
              <LevelSelector 
                selected={ this.props.severity }
                available={ this.props.allSeverities }
                onChange={ this.updateSeverity } />
             </div>
             </div>
              <div className="field is-horizontal">
              <div className="field-label is-normal">
                <label className="label">Confidence</label>
              </div>
              <div className="field-body">
              <LevelSelector
                selected={ this.props.confidence }
                available={ this.props.allConfidences }
                onChange={ this.updateConfidence } />
                </div>
              </div>
              <div className="field is-horizontal">
              <div className="field-label is-normal">
                <label className="label">Issue type</label>
              </div>
              <div className="field-body">
                <div className="field is-narrow">
                  <div className="control">
                    <div className="select is-fullwidth">
                      <select onChange={ this.updateIssueType }>
                        <option value="all" selected={ !this.props.issueType }>
                          (all)
                        </option>
                        { issueTypes }
                      </select>
                    </div>
                  </div>
                  </div>
                </div>
              </div>
              </div>
            </div>
          </nav>
        );
      }
    });
    var IssueBrowser = React.createClass({
      getInitialState: function() {
        return {};
      },
      componentWillMount: function() {
        this.updateIssues(this.props.data);
      },
      handleSeverity: function(val) {
        this.updateIssueTypes(this.props.data.Issues, val, this.state.confidence);
        this.setState({severity: val});
      },
      handleConfidence: function(val) {
        this.updateIssueTypes(this.props.data.Issues, this.state.severity, val);
        this.setState({confidence: val});
      },
      handleIssueType: function(val) {
        this.setState({issueType: val});
      },
      updateIssues: function(data) {
        if (!data) {
          this.setState({data: data});
          return;
        }
        var allSeverities = data.Issues
          .map(function(issue) {
            return issue.severity
          })
          .sort()
          .filter(function(item, pos, ary) {
            return !pos || item != ary[pos - 1];
          });
        var allConfidences = data.Issues
          .map(function(issue) {
            return issue.confidence
          })
          .sort()
          .filter(function(item, pos, ary) {
            return !pos || item != ary[pos - 1];
          });
        var selectedSeverities = allSeverities;
        var selectedConfidences = allConfidences;
        this.updateIssueTypes(data.Issues, selectedSeverities, selectedConfidences);
        this.setState({
          data: data,
          severity: selectedSeverities,
          allSeverities: allSeverities,
          confidence: selectedConfidences,
          allConfidences: allConfidences,
          issueType: null
        });
      },
      updateIssueTypes: function(issues, severities, confidences) {
        var allTypes = issues
          .filter(function(issue) {
            return severities.includes(issue.severity);
          })
          .filter(function(issue) {
            return confidences.includes(issue.confidence);
          })
          .map(function(issue) {
            return issue.details;
          })
          .sort()
          .filter(function(item, pos, ary) {
            return !pos || item != ary[pos - 1];
          });
        if (this.state.issueType && !allTypes.includes(this.state.issueType)) {
          this.setState({issueType: null});
        }
        this.setState({allIssueTypes: allTypes});
      },
      render: function() {
        return (
          <div className="content">
            <div className="columns">
              <div className="column is-one-third">
                <Navigation
                  severity={ this.state.severity } 
                  confidence={ this.state.confidence }
                  issueType={ this.state.issueType }
                  allSeverities={ this.state.allSeverities } 
                  allConfidences={ this.state.allConfidences }
                  allIssueTypes={ this.state.allIssueTypes }
                  onSeverity={ this.handleSeverity } 
                  onConfidence={ this.handleConfidence } 
                  onIssueType={ this.handleIssueType }
                />
              </div>
              <div className="column is-two-thirds">
                <Issues
                  data={ this.props.data }
                  severity={ this.state.severity }
                  confidence={ this.state.confidence }
                  issueType={ this.state.issueType }
                />
              </div>
            </div>
          </div>
        );
      }
    });
    ReactDOM.render(
      <IssueBrowser data={ data } />,
      document.getElementById("content")
    );
    hljs.highlightAll();
  </script>
</body>
</html>`
