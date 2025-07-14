# Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# -*- coding: utf-8 -*-
import logging
import threading
import sys

class thread_wrapper(threading.Thread):
    """A wrapper object to run thread functions and collect return values """
    def __init__(self, thread_name, func_obj, *args, **kargs):
        """Init method
        :param thread_name: unique name given to the thread
        :param func_obj: thread function object
        :param args: positional arguments to the thread function
        :param kargs: keyword arguments to the thread function
        """
        super(thread_wrapper, self).__init__()
        self._args = tuple(arg for arg in args)
        self._kargs = {k:v for k,v in kargs.items()}
        self._func_obj = func_obj
        self._detail = None
        self._thread_name = thread_name
        
    def run(self):
        """Overridden from threading.Thread. Called when thread_obj.start() is
        invoked. This inturn invokes the thread function and stores the return
        value in the object variable"""
        self._detail = self._func_obj(*self._args, **self._kargs)
        
    @property
    def detail(self):
        """Return value of the thread function"""
        return self._detail
    
    @property
    def thread_name(self):
        """Thread name"""
        return self._thread_name
    
    
def dispatch_and_join(thread_objs):
    """Dispatches the thread objects, joins, and returns the return values of 
    all the thread functions as a list object"""    
    for t_obj in thread_objs:
        t_obj.start()
                            
    details = []
    for t_obj in thread_objs:
        try:
            t_obj.join()
        except:
            info = sys.exc_info()
            logging.error("exception during join() of {} - {}".format
                          (t_obj.thread_name, info[0]))
        finally:                            
            if t_obj.isAlive():
                logging.error(
                        "Thread fetching {} timed out".format(
                                t_obj.thread_name))
            else:
                detail = t_obj.detail
                details.append(detail)
    return details
