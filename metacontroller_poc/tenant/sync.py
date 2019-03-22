#!/usr/bin/env python

# Copyright 2019 Google Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

from http.server import BaseHTTPRequestHandler, HTTPServer

import copy
import jinja2
import json
import os
import re
import yaml

def is_job_finished(job):
  if 'status' in job:
    status_phase = job['status'].get('phase', "NO_STATUS_PAHSE_YET")
    if status_phase == "Succeeded":
      return True
    """
    desiredNumberScheduled = job['status'].get('desiredNumberScheduled',1)
    numberReady = job['status'].get('numberReady',0)
    if desiredNumberScheduled == numberReady and desiredNumberScheduled > 0:
      return True
    """
  return False

def new_workflow(job):

  wf = {}
  template_filename = 'templates/template.j2'
  script_path = os.path.dirname(os.path.abspath(__file__))
  template_file_path = os.path.join(script_path, template_filename)
  environment = jinja2.Environment(loader=jinja2.FileSystemLoader(script_path))

  wf_text = environment.get_template(template_filename).render(job)
  wf = yaml.load(wf_text)

  return wf


class Controller(BaseHTTPRequestHandler):
  def sync(self, job, children):
    desired_status = {}
    child = '%s-dj' % (job['metadata']['name'])
    # import pdb; pdb.set_trace()

    self.log_message(" Job: %s", job)
    self.log_message(" Children: %s", children)

    # If the job already finished at some point, freeze the status,
    # delete children, and take no further action.
    if is_job_finished(job):
      desired_status = copy.deepcopy(job['status'])
      desired_status['conditions'] = [{'type': 'Complete', 'status': 'True'}]
      return {'status': desired_status, 'children': []}

    # Compute status based on what we observed, before building desired state.
    # Our .status is just a copy of the Argo Workflow .status with extra fields.
    desired_status = copy.deepcopy(children['Workflow.argoproj.io/v1alpha1'].get(child, {}).get('status',{}))
    if is_job_finished(children['Workflow.argoproj.io/v1alpha1'].get(child, {})):
      desired_status['conditions'] = [{'type': 'Complete', 'status': 'True'}]
    else:
      desired_status['conditions'] = [{'type': 'Complete', 'status': 'False'}]

    # Always generate desired state for child if we reach this point.
    # We should not delete children until after we know we've recorded
    # completion in our status, which was the first check we did above.
    desired_child = new_workflow(job)
    self.log_message(" Workflow: %s", desired_child)
    return {'status': desired_status, 'children': [desired_child]}


  def do_POST(self):
    content_in_bytes = self.rfile.read(int(self.headers.get('content-length')))
    observed = json.loads(content_in_bytes.decode('utf-8'))
    desired = self.sync(observed['parent'], observed['children'])

    self.send_response(200)
    self.send_header('Content-type', 'application/json')
    self.end_headers()
    self.wfile.write(json.dumps(desired).encode('utf-8'))

HTTPServer(('', 80), Controller).serve_forever()
