import logging
import numpy as np
import os
import tensorflow as tf
from flask import Flask, request, jsonify

#change it to local
#tf_model_path = '~/go/src/github.com/filipecosta90/aibench/tests/models/tensorflow/creditcardfraud.pb'
tf_model_path = os.getenv('TF_MODEL_PATH', '/root/data/creditcardfraud.pb' )

with tf.io.gfile.GFile(tf_model_path, "rb") as f:
    restored_graph_def = tf.compat.v1.GraphDef()
    restored_graph_def.ParseFromString(f.read())

with tf.Graph().as_default() as graph:
    tf.import_graph_def(
        restored_graph_def,
        input_map=None,
        return_elements=None,
        name="")


app = Flask(__name__)

if __name__ != '__main__':
    gunicorn_logger = logging.getLogger('gunicorn.error')
    app.logger.handlers = gunicorn_logger.handlers
    app.logger.setLevel(gunicorn_logger.level)

app.logger.info('reading model from {0}'.format(tf_model_path))

app.logger.info(os.path.isfile(tf_model_path))

sess = tf.compat.v1.Session(graph=graph)
transaction_tensor = graph.get_tensor_by_name('transaction:0')
reference_tensor = graph.get_tensor_by_name('reference:0')
output_tensor = graph.get_tensor_by_name('output:0')

app.logger.info('model read from {0}'.format(tf_model_path))


@app.route('/predict', methods=['POST'])
def predict():
    transaction_data = np.array(request.json['inputs']['transaction'], dtype=np.float32)
    reference_data = np.array(request.json['inputs']['reference'], dtype=np.float32)
    out = sess.run(output_tensor, feed_dict={transaction_tensor: transaction_data, reference_tensor: reference_data})
    response = {'outputs': out.tolist()}
    return jsonify(response)


if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8000, debug=True)