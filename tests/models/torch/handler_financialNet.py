# custom service file

# model_handler.py

# https://pytorch.org/serve/custom_service.html
# https://pytorch.org/serve/logging.html
# https://pytorch.org/serve/server.html

# torch-model-archiver --model-name financialNetTorch --version 1 --serialized-file torchFraudNetWithRef.pt --handler handler_financialNet.py
# torchserve --start --model-store . --models financial=financialNetTorch.mar --ts-config config.properties --log-config log4j.properties
"""
ModelHandler defines a base model handler.
"""
import io
import logging
import numpy as np
import os
import torch
import json
import array

logger = logging.getLogger(__name__)

class ModelHandler(object):
    """
    A base Model handler implementation.
    """

    def __init__(self):
        self.error = None
        self._context = None
        self.model=None
        self._batch_size = 0
        self.device = None
        self.initialized = False

    def initialize(self, context):
        """
        Initialize model. This will be called during model loading time
        :param context: Initial context contains model server system properties.
        :return:
        """
        self._context = context
        properties = context.system_properties
        self._batch_size = properties["batch_size"]
        self.device = torch.device("cuda:" + str(properties.get("gpu_id")) if torch.cuda.is_available() else "cpu")
        model_dir = properties.get("model_dir")
        # Read model serialize/pt file
        model_pt_path = os.path.join(model_dir, "torchFraudNetWithRef.pt")
        self.model = model = torch.jit.load(model_pt_path)
        self.initialized = True

    def preprocess(self, batch):
        """
        Transform raw input into model input data.
        :param batch: list of raw requests, should match batch size
        :return: list of preprocessed model input data
        """
        # Take the input data and pre-process it make it inference ready
#         assert self._batch_size == len(batch), "Invalid input batch size: {}".format(len(batch))
        return batch

    def inference(self, model_input):
        response = {'outputs': None }
        """
        Internal inference methods, checks if the input data has the correct format
        :param model_input: transformed model input data
        :return: list of inference output
        """
        if 'body' in model_input[0]:
            body = model_input[0]['body']
            if 'transaction' in body and 'reference' in body:
                transaction_data = np.array(body['transaction'], dtype=np.float32).reshape(1, 30)
                reference_data = np.array(body['reference'], dtype=np.float32)
                torch_tensor1 = torch.from_numpy(transaction_data)
                torch_tensor2 = torch.from_numpy(reference_data)
                with torch.no_grad():
                    out = self.model(torch_tensor1,torch_tensor2).numpy().tolist()
                    response['outputs']=out[0]
        return response

    def postprocess(self, inference_output):
        # Take output from network and post-process to desired format
        return [inference_output]

    def handle(self, data, context):
        model_input = self.preprocess(data)
        model_out = self.inference(model_input)
        return self.postprocess(model_out)

_service = ModelHandler()


def handle(data, context):
    if not _service.initialized:
        _service.initialize(context)

    if data is None:
        return None

    return _service.handle(data, context)